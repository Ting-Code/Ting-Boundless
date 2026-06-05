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
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "auth-service"

func main() {
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	health := httpx.NewHealth()

	mux := http.NewServeMux()
	health.Handler(mux)

	// Logto webhook -> verify signature, normalize, idempotency, emit audit.
	mux.HandleFunc("POST /internal/webhooks/logto", handleLogtoWebhook)
	// Mini-program login: exchange wx code (code2session) -> issue standard JWT.
	mux.HandleFunc("POST /v1/auth/miniprogram/login", handleMiniProgramLogin)

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

func handleLogtoWebhook(w http.ResponseWriter, _ *http.Request) {
	// TODO: verify Logto signature, dedupe by event id, map to audit.Event,
	// deliver to Audit Service.
	httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func handleMiniProgramLogin(w http.ResponseWriter, _ *http.Request) {
	// TODO: call WeChat code2session with WECHAT_APP_ID/SECRET, resolve openid,
	// upsert user, then issue a standard JWT (issuer + JWKS the Gateway trusts).
	httpx.JSON(w, http.StatusNotImplemented, map[string]string{
		"code":    "not_implemented",
		"message": "miniprogram login not implemented",
	})
}
