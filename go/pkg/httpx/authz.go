package httpx

import (
	"net/http"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/identity"
)

// RequireAuthenticated rejects requests without a trusted user id in context.
// Compose after identity.Middleware.
func RequireAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := identity.FromContext(r.Context())
		if !ok || !id.Authenticated() {
			WriteError(w, id.RequestID, errs.Unauthorized("unauthenticated", "authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole rejects requests unless the actor has the given role.
// Compose after identity.Middleware (and usually RequireAuthenticated).
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := identity.FromContext(r.Context())
			if !ok || !id.Authenticated() {
				WriteError(w, id.RequestID, errs.Unauthorized("unauthenticated", "authentication required"))
				return
			}
			if !identity.HasRole(id, role) {
				WriteError(w, id.RequestID, errs.Forbidden("forbidden", role+" role required"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
