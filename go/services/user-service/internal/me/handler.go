package me

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/user-service/internal/store"
)

const maxDisplayNameLen = 200

type userStore interface {
	GetOrCreate(ctx context.Context, id identity.Identity) (store.User, error)
	UpdateDisplayName(ctx context.Context, id identity.Identity, displayName string) (store.User, error)
}

// Handler serves GET and PATCH /v1/users/me.
type Handler struct {
	users userStore
}

// New returns a me handler.
func New(users userStore) http.Handler {
	return &Handler{users: users}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPatch:
		h.patch(w, r)
	default:
		id, _ := identity.FromContext(r.Context())
		httpx.WriteError(w, id.RequestID, errs.BadRequest("method_not_allowed", "use GET or PATCH"))
	}
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.WriteError(w, id.RequestID, errs.Unauthorized("unauthenticated", "authentication required"))
		return
	}
	if h.users == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}

	u, err := h.users.GetOrCreate(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("user_lookup_failed", "failed to load user profile"))
		return
	}
	writeProfile(w, u, id)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.WriteError(w, id.RequestID, errs.Unauthorized("unauthenticated", "authentication required"))
		return
	}
	if h.users == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("invalid_body", "could not read request body"))
		return
	}
	var req struct {
		DisplayName *string `json:"display_name"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("invalid_json", "malformed JSON body"))
		return
	}
	if req.DisplayName == nil {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("display_name_required", "display_name is required"))
		return
	}
	name := strings.TrimSpace(*req.DisplayName)
	if name == "" {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("display_name_empty", "display_name cannot be empty"))
		return
	}
	if len(name) > maxDisplayNameLen {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("display_name_too_long", "display_name exceeds limit"))
		return
	}

	u, err := h.users.UpdateDisplayName(r.Context(), id, name)
	if err != nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("user_update_failed", "failed to update profile"))
		return
	}
	writeProfile(w, u, id)
}

func writeProfile(w http.ResponseWriter, u store.User, id identity.Identity) {
	httpx.JSON(w, http.StatusOK, map[string]any{
		"user_id":      u.ID,
		"tenant_id":    u.TenantID,
		"display_name": u.DisplayName,
		"roles":        id.Roles,
		"created_at":   u.CreatedAt,
		"updated_at":   u.UpdatedAt,
	})
}
