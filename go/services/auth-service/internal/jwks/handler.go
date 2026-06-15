package jwks

import (
	"net/http"

	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// Handler serves the JWKS document for Gateway verification.
type Handler struct {
	issuer *auth.Issuer
}

// NewHandler builds a JWKS endpoint handler.
func NewHandler(issuer *auth.Issuer) *Handler {
	return &Handler{issuer: issuer}
}

// ServeHTTP handles GET /v1/auth/jwks.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rid := r.Header.Get("X-Request-Id")
	if h == nil || h.issuer == nil {
		httpx.WriteError(w, rid, errs.Internal("auth.unavailable", "jwks not configured"))
		return
	}
	body, err := h.issuer.JWKSJSON()
	if err != nil {
		httpx.WriteError(w, rid, errs.Internal("auth.jwks_failed", "could not build jwks"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
