package identityapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
)

type providerResolver interface {
	ResolveProviderUser(ctx context.Context, provider, providerUID string) (string, error)
}

type resolveRequest struct {
	Provider    string `json:"provider"`
	ProviderUID string `json:"provider_uid"`
}

type resolveResponse struct {
	UserID string `json:"user_id"`
}

// ResolveHandler maps external IdP subjects to platform user IDs.
type ResolveHandler struct {
	store providerResolver
}

// NewResolveHandler builds the internal identity resolver.
func NewResolveHandler(s providerResolver) *ResolveHandler {
	return &ResolveHandler{store: s}
}

// ServeHTTP handles POST /internal/identity/resolve.
func (h *ResolveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rid := r.Header.Get(identity.HeaderRequestID)

	if h == nil || h.store == nil {
		httpx.WriteError(w, rid, errs.Internal("identity_unavailable", "identity store not configured"))
		return
	}

	var req resolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Provider == "" || req.ProviderUID == "" {
		httpx.WriteError(w, rid, errs.BadRequest("invalid_request", "provider and provider_uid are required"))
		return
	}

	userID, err := h.store.ResolveProviderUser(r.Context(), req.Provider, req.ProviderUID)
	if err != nil {
		httpx.WriteError(w, rid, errs.Internal("identity_resolve_failed", "could not resolve identity"))
		return
	}

	httpx.JSON(w, http.StatusOK, resolveResponse{UserID: userID})
}
