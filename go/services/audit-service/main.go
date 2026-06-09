// Command audit-service consumes audit events and persists them to audit_db
// (append-only). It is a platform service and owns no business domain logic.
//
// V1 ingestion is via HTTP from the outbox dispatcher / auth-service; V2 moves to
// RabbitMQ. Storage should be append-only with a retention policy.
package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/services/audit-service/internal/store"
)

const serviceName = "audit-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	auditDB := httpx.Env("AUDIT_DB", "audit_db")
	cfg := db.ConfigFromEnv(auditDB)
	pg := db.Connect(ctx, log, auditDB)
	if pg.DB != nil {
		defer pg.DB.Close()
		if err := db.RunMigrations(cfg, serviceName); err != nil {
			log.Error("migrations failed", slog.Any("error", err))
			return
		}
	}

	var events *store.Events
	if pg.DB != nil {
		events = store.NewEvents(pg.DB.Pool())
	}

	internalToken := httpx.Env("INTERNAL_API_TOKEN", "")

	health := httpx.NewHealth()
	db.RegisterHealth(health, "audit_db", pg.Probe)

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("POST /internal/audit/events",
		httpx.InternalAuth(internalToken)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleIngest(w, r, events)
			}),
		),
	)

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8085")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleIngest(w http.ResponseWriter, r *http.Request, events *store.Events) {
	rid := r.Header.Get(identity.HeaderRequestID)
	if events == nil {
		errs.Write(w, rid, errs.Internal("database_unavailable", "audit database not connected"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		errs.Write(w, rid, errs.BadRequest("invalid_body", "could not read request body"))
		return
	}

	var e audit.Event
	if err := json.Unmarshal(body, &e); err != nil {
		errs.Write(w, rid, errs.BadRequest("invalid_event", "malformed audit event"))
		return
	}
	if e.ID == "" || e.Source == "" || e.Type == "" || e.Time.IsZero() {
		errs.Write(w, rid, errs.BadRequest("invalid_event", "id, source, type, and time are required"))
		return
	}

	if err := events.Insert(r.Context(), e); err != nil {
		logger.From(r.Context()).Error("audit insert failed", slog.Any("error", err))
		errs.Write(w, rid, errs.Internal("persist_failed", "failed to persist audit event"))
		return
	}

	httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "id": e.ID})
}
