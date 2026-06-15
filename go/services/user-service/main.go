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
	"github.com/ting-boundless/boundless/services/user-service/internal/list"
	"github.com/ting-boundless/boundless/services/user-service/internal/me"
	"github.com/ting-boundless/boundless/services/user-service/internal/store"
)

const serviceName = "user-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	cfg := db.ConfigFromEnv("")
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
		if err := db.RunMigrations(cfg, serviceName); err != nil {
			log.Error("migrations failed", slog.Any("error", err))
			return
		}
	}

	var users *store.Users
	if pg.DB != nil {
		users = store.NewUsers(pg.DB.Pool())
	}

	health := httpx.NewHealth()
	db.RegisterHealth(health, "postgres", pg.Probe)

	meHandler := identity.Middleware(me.New(users))

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("/v1/users/me", meHandler)
	mux.Handle("GET /v1/users/", identity.Middleware(httpx.RequireRole("admin")(list.New(users))))

	internalToken, ok := httpx.LoadInternalToken(log)
	if !ok {
		return
	}

	h := httpx.Chain(mux,
		httpx.GatewayTrust(internalToken),
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
		httpx.TraceContext,
	)

	addr := httpx.Env("HTTP_ADDR", ":8081")
	if err := httpx.RunService(addr, serviceName, h, log); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}
