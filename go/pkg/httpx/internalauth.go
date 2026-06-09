package httpx

import (
	"net/http"
	"strings"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/identity"
)

// InternalAuth protects /internal/* routes with a shared secret token.
// Set INTERNAL_API_TOKEN in env; callers send Authorization: Bearer <token>
// or X-Internal-Token: <token>.
func InternalAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get(identity.HeaderRequestID)
			if token == "" {
				errs.Write(w, rid, errs.Internal("internal_auth_misconfigured", "internal auth not configured"))
				return
			}
			if !internalTokenOK(r, token) {
				errs.Write(w, rid, errs.Unauthorized("unauthorized", "invalid internal token"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func internalTokenOK(r *http.Request, want string) bool {
	if h := r.Header.Get("X-Internal-Token"); h == want {
		return true
	}
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ") == want
	}
	return false
}
