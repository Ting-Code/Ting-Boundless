package httpx

import (
	"net/http"
	"strings"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/identity"
)

// GatewayTrust rejects external callers that lack the shared internal token.
// Health probes and /internal/* (separate InternalAuth) are exempt.
// When token is empty (local dev), the check is skipped.
func GatewayTrust(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" || gatewayTrustSkipped(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			rid := r.Header.Get(identity.HeaderRequestID)
			if !internalTokenOK(r, token) {
				errs.Write(w, rid, errs.Unauthorized("untrusted_caller", "request must come through Gateway"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func gatewayTrustSkipped(path string) bool {
	if path == "/healthz" || path == "/readyz" || path == "/metrics" {
		return true
	}
	return strings.HasPrefix(path, "/internal/")
}
