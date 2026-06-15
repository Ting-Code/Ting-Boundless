// Command audit-service consumes audit events and persists them to audit_db
// (append-only). It is a platform service and owns no business domain logic.
//
// V1 ingestion is via HTTP from the outbox dispatcher / auth-service; V2 moves to
// RabbitMQ. Storage should be append-only with a retention policy.
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/services/audit-service/internal/ingest"
	"github.com/ting-boundless/boundless/services/audit-service/internal/query"
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

	internalToken, ok := httpx.LoadInternalToken(log)
	if !ok {
		return
	}

	health := httpx.NewHealth()
	db.RegisterHealth(health, "audit_db", pg.Probe)

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("POST /internal/audit/events",
		httpx.InternalAuth(internalToken)(ingest.New(events)),
	)
	mux.Handle("GET /v1/audit/events", identity.Middleware(query.New(events)))

	h := httpx.Chain(mux,
		httpx.GatewayTrust(internalToken),
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
		httpx.TraceContext,
	)

	addr := httpx.Env("HTTP_ADDR", ":8085")
	if err := httpx.RunService(addr, serviceName, h, log); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}
