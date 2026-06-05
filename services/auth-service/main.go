// Command auth-service is the IdP integration layer.
//
// Responsibilities (see docs/ARCHITECTURE.md "Auth Service / IdP Integration"):
//   - receive and verify Logto webhooks (signature, normalization, idempotency)
//   - convert identity events into CloudEvents audit events
//   - WeChat mini-program login: code2session, then issue a STANDARD JWT
//   - optionally act as an internal OIDC issuer (own issuer + JWKS)
//
// Any token minted here must be a standard JWT verifiable by the Gateway via a
// known issuer + JWKS. Never mint ad-hoc/custom tokens.
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/cache"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "auth-service"

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
	mux.HandleFunc("POST /internal/webhooks/logto", handleLogtoWebhook)
	mux.HandleFunc("POST /v1/auth/miniprogram/login", handleMiniProgramLogin)

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8084")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleLogtoWebhook(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func handleMiniProgramLogin(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusNotImplemented, map[string]string{
		"code":    "not_implemented",
		"message": "miniprogram login not implemented",
	})
}
