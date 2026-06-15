package query

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/audit-service/internal/store"
)

type eventStore interface {
	List(ctx context.Context, f store.ListFilter) ([]store.EventRow, error)
}

// Handler serves GET /v1/audit/events.
type Handler struct {
	events eventStore
}

// New returns the audit query handler.
func New(events eventStore) http.Handler {
	return &Handler{events: events}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		id, _ := identity.FromContext(r.Context())
		httpx.WriteError(w, id.RequestID, errs.BadRequest("method_not_allowed", "use GET"))
		return
	}

	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.WriteError(w, id.RequestID, errs.Unauthorized("unauthenticated", "authentication required"))
		return
	}
	if !identity.HasRole(id, "admin") {
		httpx.WriteError(w, id.RequestID, errs.Forbidden("forbidden", "admin role required"))
		return
	}
	if h.events == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "audit database not connected"))
		return
	}

	q := r.URL.Query()
	limit := 50
	if raw := q.Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}

	events, err := h.events.List(r.Context(), store.ListFilter{
		TenantID: id.TenantID,
		Type:     q.Get("type"),
		Source:   q.Get("source"),
		Limit:    limit,
	})
	if err != nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("query_failed", "failed to list audit events"))
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"events": events})
}
