// Package auth implements Gateway edge authentication and identity injection.
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/revocation"
	"github.com/ting-boundless/boundless/services/gateway/internal/bff"
	"github.com/ting-boundless/boundless/services/gateway/internal/identityresolve"
	"github.com/ting-boundless/boundless/services/gateway/internal/session"
)

// Authenticate strips untrusted headers, assigns a fresh request_id, verifies
// Bearer JWT or BFF session cookie, and injects trusted identity headers.
//
// Paths matching anonPrefixes may proceed without credentials; all other paths
// are rejected at the Gateway when no valid token or session is present.
func Authenticate(v *auth.Verifier, sessions *session.Store, anon AnonPrefixes, resolver *identityresolve.Client, entryAudit audit.Emitter, revocations *revocation.Store, sensitive SensitivePrefixes) func(http.Handler) http.Handler {
	if anon.IsEmpty() {
		anon = DefaultAnonPrefixes()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity.StripUntrusted(r.Header)

			rid := httpx.NewRequestID()
			r.Header.Set(identity.HeaderRequestID, rid)
			w.Header().Set(identity.HeaderRequestID, rid)

			id := identity.Identity{RequestID: rid}
			authenticated := false
			sessionID := ""

			if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {
				if v == nil || !v.Enabled() {
					httpx.WriteError(w, rid, errs.Internal("auth_unconfigured", "jwt verification not configured"))
					return
				}
				raw := strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer "))
				verified, err := v.Verify(raw)
				if err != nil {
					audit.EmitEntry(entryAudit, r, rid, "api.token.invalid", map[string]any{
						"reason": "invalid_token",
					})
					httpx.WriteError(w, rid, errs.Unauthorized("invalid_token", "invalid or expired token"))
					return
				}
				id = verified
				id.RequestID = rid
				id = mapPlatformUserID(r.Context(), resolver, id, rid)
				authenticated = id.UserID != "" || id.Subject != ""
			} else if sessions != nil && sessions.Enabled() && v != nil && v.Enabled() {
				if cookieID, sid, err := bff.IdentityFromRequest(r, sessions, v); err == nil {
					sessionID = sid
					cookieID.RequestID = rid
					id = mapPlatformUserID(r.Context(), resolver, cookieID, rid)
					authenticated = id.UserID != "" || id.Subject != ""
				}
			}

			if !authenticated && !anon.Allows(r.URL.Path) {
				audit.EmitEntry(entryAudit, r, rid, "api.access.denied", map[string]any{
					"reason": "auth.unauthenticated",
				})
				httpx.WriteError(w, rid, errs.Unauthorized("auth.unauthenticated", "authentication required"))
				return
			}

			if authenticated && sensitive.RequiresRevocationCheck(r.URL.Path) {
				if denied, reason := checkRevocation(r.Context(), revocations, id, sessionID); denied {
					audit.EmitEntry(entryAudit, r, rid, "api.access.denied", map[string]any{
						"reason": reason,
					})
					httpx.WriteError(w, rid, errs.Unauthorized("auth.revoked", "credentials have been revoked"))
					return
				}
			}

			id.Inject(r.Header)
			next.ServeHTTP(w, r)
		})
	}
}

func checkRevocation(ctx context.Context, revocations *revocation.Store, id identity.Identity, sessionID string) (denied bool, reason string) {
	if revocations == nil || !revocations.Enabled() {
		return false, ""
	}
	if sessionID != "" {
		revoked, err := revocations.IsSessionRevoked(ctx, sessionID)
		if err == nil && revoked {
			return true, "auth.session_revoked"
		}
	}
	subject := id.Subject
	if subject == "" {
		subject = id.UserID
	}
	if subject != "" {
		revoked, err := revocations.IsSubjectRevoked(ctx, subject)
		if err == nil && revoked {
			return true, "auth.subject_revoked"
		}
	}
	return false, ""
}
