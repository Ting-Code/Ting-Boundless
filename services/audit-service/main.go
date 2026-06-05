// Command audit-service consumes audit events and persists them to audit_db
// (append-only). It is a platform service and owns no business domain logic.
//
// V1 ingestion is via HTTP from the outbox dispatcher / auth-service; V2 moves to
// RabbitMQ. Storage should be append-only with a retention policy.
package main

import (
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "audit-service"

func main() {
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	health := httpx.NewHealth()
	// TODO: health.Register(httpx.Check{Name: "audit_db", Probe: db.Ping})

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.HandleFunc("POST /internal/audit/events", handleIngest(log))

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8080")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleIngest(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var _ audit.Event
		// TODO: decode CloudEvents-style event, validate against schema,
		// idempotency by event id, INSERT append-only into audit_db.
		logger.From(r.Context()).Info("audit event received")
		httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
	}
}
