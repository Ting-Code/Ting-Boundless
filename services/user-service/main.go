// Command user-service owns the user domain.
//
// It NEVER parses end-user JWTs. It trusts the identity context injected by the
// Gateway (identity.Middleware) and enforces domain authorization locally.
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
)

const serviceName = "user-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
	}

	health := httpx.NewHealth()
	db.RegisterHealth(health, "postgres", pg.Probe)

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("GET /v1/users/me", identity.Middleware(http.HandlerFunc(handleMe)))

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8081")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.JSON(w, http.StatusUnauthorized, map[string]string{"code": "unauthenticated"})
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"user_id":   id.UserID,
		"tenant_id": id.TenantID,
		"roles":     id.Roles,
	})
}
