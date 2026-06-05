// Command business-service owns a core business domain.
//
// Domain authorization (resource ownership, tenant isolation, business-state
// rules) lives here. Identity comes from the Gateway via identity.Middleware.
package main

import (
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "business-service"

func main() {
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	health := httpx.NewHealth()

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("GET /v1/business/ping", identity.Middleware(http.HandlerFunc(handlePing)))

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

func handlePing(w http.ResponseWriter, r *http.Request) {
	id, _ := identity.FromContext(r.Context())
	// Example: enforce tenant isolation here before touching data.
	httpx.JSON(w, http.StatusOK, map[string]any{"pong": true, "tenant_id": id.TenantID})
}
