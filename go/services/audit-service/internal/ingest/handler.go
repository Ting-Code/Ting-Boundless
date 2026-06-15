package ingest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/contracts"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
)

type eventWriter interface {
	Insert(ctx context.Context, e audit.Event) error
}

// Handler serves POST /internal/audit/events.
type Handler struct {
	events eventWriter
}

// New returns an ingest handler.
func New(events eventWriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, events)
	})
}

func handle(w http.ResponseWriter, r *http.Request, events eventWriter) {
	rid := r.Header.Get(identity.HeaderRequestID)
	if events == nil {
		httpx.WriteError(w, rid, errs.Internal("database_unavailable", "audit database not connected"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.WriteError(w, rid, errs.BadRequest("invalid_body", "could not read request body"))
		return
	}

	var e audit.Event
	if err := json.Unmarshal(body, &e); err != nil {
		httpx.WriteError(w, rid, errs.BadRequest("invalid_event", "malformed audit event"))
		return
	}
	if e.ID == "" || e.Source == "" || e.Type == "" || e.Time.IsZero() {
		httpx.WriteError(w, rid, errs.BadRequest("invalid_event", "id, source, type, and time are required"))
		return
	}

	if _, err := contracts.AuditToProto(e); err != nil {
		httpx.WriteError(w, rid, errs.BadRequest("invalid_event_data", "event data is not contract-compatible"))
		return
	}

	if err := events.Insert(r.Context(), e); err != nil {
		logger.From(r.Context()).Error("audit insert failed", slog.Any("error", err))
		httpx.WriteError(w, rid, errs.Internal("persist_failed", "failed to persist audit event"))
		return
	}

	httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "id": e.ID})
}
