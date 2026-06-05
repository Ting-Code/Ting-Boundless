// Command business-service owns a core business domain.
//
// Domain authorization (resource ownership, tenant isolation, business-state
// rules) lives here. Identity comes from the Gateway via identity.Middleware.
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/cache"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "business-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
	}
	rd := cache.Connect(ctx, log)
	if rd.Client != nil {
		defer rd.Client.Close()
	}

	health := httpx.NewHealth()
	db.RegisterHealth(health, "postgres", pg.Probe)
	cache.RegisterHealth(health, rd.Probe)

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("GET /v1/business/ping", identity.Middleware(http.HandlerFunc(handlePing)))

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8082")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	id, _ := identity.FromContext(r.Context())
	httpx.JSON(w, http.StatusOK, map[string]any{"pong": true, "tenant_id": id.TenantID})
}
