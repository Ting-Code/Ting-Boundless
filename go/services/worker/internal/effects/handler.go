package effects

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/worker/internal/store"
)

type effectLister interface {
	List(ctx context.Context, f store.ListFilter) ([]store.EffectRow, error)
}

// Handler serves GET /internal/job-effects.
type Handler struct {
	effects effectLister
}

// New returns the internal job-effects list handler.
func New(effects effectLister) http.Handler {
	return &Handler{effects: effects}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		rid := r.Header.Get(identity.HeaderRequestID)
		httpx.WriteError(w, rid, errs.BadRequest("method_not_allowed", "use GET"))
		return
	}
	rid := r.Header.Get(identity.HeaderRequestID)
	if h.effects == nil {
		httpx.WriteError(w, rid, errs.Internal("database_unavailable", "database not connected"))
		return
	}

	q := r.URL.Query()
	limit := 50
	if raw := q.Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}

	rows, err := h.effects.List(r.Context(), store.ListFilter{
		TenantID: q.Get("tenant_id"),
		JobType:  q.Get("job_type"),
		Limit:    limit,
	})
	if err != nil {
		httpx.WriteError(w, rid, errs.Internal("query_failed", "failed to list job effects"))
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"effects": rows})
}
