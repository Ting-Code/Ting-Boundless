// Package identity models the trusted identity context that the Gateway injects
// and that downstream services consume.
//
// Rules (see docs/AI_CONTEXT.md):
//   - The Gateway is the only component that may set these headers from verified
//     tokens. It MUST strip any client-supplied values first (StripUntrusted).
//   - Business services NEVER parse end-user JWTs; they read the identity from
//     these headers (FromHeaders) and trust them only on internal traffic.
package identity

import (
	"context"
	"net/http"
	"strings"
)

// Trusted identity headers. Keep in sync with
// platform-contracts/schemas/identity-context.schema.json.
const (
	HeaderRequestID   = "X-Request-Id"
	HeaderUserID      = "X-User-Id"
	HeaderTenantID    = "X-Tenant-Id"
	HeaderRoles       = "X-Roles"       // comma-separated
	HeaderScopes      = "X-Scopes"      // comma-separated
	HeaderAuthSubject = "X-Auth-Subject"
)

// allHeaders is the full set the Gateway strips from external requests.
var allHeaders = []string{
	HeaderUserID, HeaderTenantID, HeaderRoles, HeaderScopes, HeaderAuthSubject, HeaderRequestID,
}

// Identity is the actor context carried with every request and internal call.
type Identity struct {
	RequestID string
	UserID    string
	TenantID  string
	Roles     []string
	Scopes    []string
	Subject   string
}

// StripUntrusted removes any client-supplied identity headers. The Gateway calls
// this on inbound external requests before injecting verified values.
func StripUntrusted(h http.Header) {
	for _, k := range allHeaders {
		h.Del(k)
	}
}

// Inject writes the identity onto outgoing headers (Gateway -> service, or
// service -> service propagation).
func (id Identity) Inject(h http.Header) {
	setNonEmpty(h, HeaderRequestID, id.RequestID)
	setNonEmpty(h, HeaderUserID, id.UserID)
	setNonEmpty(h, HeaderTenantID, id.TenantID)
	setNonEmpty(h, HeaderRoles, strings.Join(id.Roles, ","))
	setNonEmpty(h, HeaderScopes, strings.Join(id.Scopes, ","))
	setNonEmpty(h, HeaderAuthSubject, id.Subject)
}

// FromHeaders reads the identity from trusted (internal) request headers.
func FromHeaders(h http.Header) Identity {
	return Identity{
		RequestID: h.Get(HeaderRequestID),
		UserID:    h.Get(HeaderUserID),
		TenantID:  h.Get(HeaderTenantID),
		Roles:     splitCSV(h.Get(HeaderRoles)),
		Scopes:    splitCSV(h.Get(HeaderScopes)),
		Subject:   h.Get(HeaderAuthSubject),
	}
}

type ctxKey struct{}

// NewContext stores the identity in the context.
func NewContext(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext retrieves the identity from the context.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// Middleware extracts the identity from trusted headers into the context. Use it
// in business services, which must only see internal Gateway/service traffic.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r.Context(), FromHeaders(r.Header))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setNonEmpty(h http.Header, k, v string) {
	if v != "" {
		h.Set(k, v)
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
