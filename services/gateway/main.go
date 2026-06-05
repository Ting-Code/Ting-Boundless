// Command gateway is the edge API entry (Go Gateway/BFF).
//
// Responsibilities (see docs/ARCHITECTURE.md "Go Gateway / BFF"):
//   - verify caller credentials (Bearer JWT via cached JWKS; cookie for web BFF)
//   - STRIP client-supplied identity headers, then inject trusted ones
//   - route to business services, apply rate limiting, emit entry audit events
//
// Business authorization does NOT live here. The Gateway decides "who you are";
// services decide "what you may do".
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/cache"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/services/gateway/internal/proxy"
)

const serviceName = "gateway"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	rd := cache.Connect(ctx, log)
	if rd.Client != nil {
		defer rd.Client.Close()
	}

	health := httpx.NewHealth()
	cache.RegisterHealth(health, rd.Probe)

	routes := proxy.Routes{
		"/v1/users/":    httpx.Env("USER_SERVICE_URL", "http://127.0.0.1:8081"),
		"/v1/business/": httpx.Env("BUSINESS_SERVICE_URL", "http://127.0.0.1:8082"),
		"/v1/files/":    httpx.Env("FILE_SERVICE_URL", "http://127.0.0.1:8083"),
	}

	router, err := proxy.New(routes, log)
	if err != nil {
		log.Error("failed to build proxy", slog.Any("error", err))
		return
	}

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("/", router)

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
		authenticate(),
	)

	addr := httpx.Env("HTTP_ADDR", ":8080")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func authenticate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity.StripUntrusted(r.Header)
			id := identity.Identity{RequestID: r.Header.Get(identity.HeaderRequestID)}
			id.Inject(r.Header)
			next.ServeHTTP(w, r)
		})
	}
}
