// Package auth implements Gateway edge authentication and identity injection.
package auth

import (
	"net/http"
	"strings"

	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/trace"
	"github.com/ting-boundless/boundless/services/gateway/internal/bff"
	"github.com/ting-boundless/boundless/services/gateway/internal/session"
)

// Authenticate strips untrusted headers, assigns a fresh request_id, verifies
// Bearer JWT or BFF session cookie, and injects trusted identity headers.
//
// Paths matching anonPrefixes may proceed without credentials; all other paths
// are rejected at the Gateway when no valid token or session is present.
func Authenticate(v *auth.Verifier, sessions *session.Store, anon AnonPrefixes) func(http.Handler) http.Handler {
	if len(anon) == 0 {
		anon = DefaultAnonPrefixes()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity.StripUntrusted(r.Header)

			rid := httpx.NewRequestID()
			r.Header.Set(identity.HeaderRequestID, rid)
			w.Header().Set(identity.HeaderRequestID, rid)
			trace.EnsureRequest(r.Header, w.Header().Set)

			id := identity.Identity{RequestID: rid}
			authenticated := false

			if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {
				if v == nil || !v.Enabled() {
					errs.Write(w, rid, errs.Internal("auth_unconfigured", "jwt verification not configured"))
					return
				}
				raw := strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer "))
				verified, err := v.Verify(raw)
				if err != nil {
					errs.Write(w, rid, errs.Unauthorized("invalid_token", "invalid or expired token"))
					return
				}
				id = verified
				id.RequestID = rid
				authenticated = id.UserID != "" || id.Subject != ""
			} else if sessions != nil && sessions.Enabled() && v != nil && v.Enabled() {
				if cookieID, err := bff.IdentityFromRequest(r, sessions, v); err == nil {
					cookieID.RequestID = rid
					id = cookieID
					authenticated = id.UserID != "" || id.Subject != ""
				}
			}

			if !authenticated && !anon.Allows(r.URL.Path) {
				errs.Write(w, rid, errs.Unauthorized("auth.unauthenticated", "authentication required"))
				return
			}

			id.Inject(r.Header)
			next.ServeHTTP(w, r)
		})
	}
}
