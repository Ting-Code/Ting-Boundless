package list

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/user-service/internal/store"
)

type userLister interface {
	ListByTenant(ctx context.Context, tenantID string, limit int) ([]store.User, error)
}

// Handler serves GET /v1/users/ (tenant directory for admins).
type Handler struct {
	users userLister
}

// New returns a tenant user list handler.
func New(users userLister) http.Handler {
	return &Handler{users: users}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		id, _ := identity.FromContext(r.Context())
		httpx.WriteError(w, id.RequestID, errs.BadRequest("method_not_allowed", "use GET"))
		return
	}

	id, _ := identity.FromContext(r.Context())
	if h.users == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}

	rows, err := h.users.ListByTenant(r.Context(), id.TenantID, limit)
	if err != nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("user_list_failed", "failed to list users"))
		return
	}

	users := make([]map[string]any, 0, len(rows))
	for _, u := range rows {
		users = append(users, map[string]any{
			"user_id":      u.ID,
			"tenant_id":    u.TenantID,
			"display_name": u.DisplayName,
			"created_at":   u.CreatedAt,
			"updated_at":   u.UpdatedAt,
		})
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"users": users})
}
