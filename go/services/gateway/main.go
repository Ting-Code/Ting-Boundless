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

	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/cache"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/oidc"
	"github.com/ting-boundless/boundless/services/gateway/internal/adminstatic"
	gwauth "github.com/ting-boundless/boundless/services/gateway/internal/auth"
	"github.com/ting-boundless/boundless/services/gateway/internal/bff"
	"github.com/ting-boundless/boundless/services/gateway/internal/proxy"
	"github.com/ting-boundless/boundless/services/gateway/internal/session"
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

	authCfg := auth.ConfigFromEnv()
	verifier, err := auth.NewVerifier(ctx, authCfg, log)
	if err != nil {
		log.Error("jwt verifier init failed", slog.Any("error", err))
		return
	}

	oidcCfg := oidc.ConfigFromEnv()
	sessions := session.NewStore(rd.Client)
	bffHandler := bff.New(oidcCfg, sessions, verifier, authCfg, log)

	if oidcCfg.Ready() {
		log.Info("oidc bff enabled", slog.String("redirect_uri", oidcCfg.RedirectURI))
	} else {
		log.Warn("oidc bff disabled (set OIDC_CLIENT_ID and OIDC_CLIENT_SECRET for Logto)")
	}

	health := httpx.NewHealth()
	cache.RegisterHealth(health, rd.Probe)

	internalToken := httpx.Env("INTERNAL_API_TOKEN", "")

	routes := proxy.Routes{
		"/v1/users/":    httpx.Env("USER_SERVICE_URL", "http://127.0.0.1:8081"),
		"/v1/business/": httpx.Env("BUSINESS_SERVICE_URL", "http://127.0.0.1:3005"),
		"/v1/files/":    httpx.Env("FILE_SERVICE_URL", "http://127.0.0.1:8083"),
		"/v1/auth/":     httpx.Env("AUTH_SERVICE_URL", "http://127.0.0.1:8084"),
	}
	log.Info("upstream routes",
		slog.String("users", routes["/v1/users/"]),
		slog.String("business", routes["/v1/business/"]),
		slog.String("files", routes["/v1/files/"]),
		slog.String("auth", routes["/v1/auth/"]),
	)

	router, err := proxy.New(routes, log, internalToken)
	if err != nil {
		log.Error("failed to build proxy", slog.Any("error", err))
		return
	}

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.HandleFunc("GET /sign-in", bffHandler.SignIn)
	mux.HandleFunc("GET /callback", bffHandler.Callback)
	mux.HandleFunc("GET /sign-out", bffHandler.SignOut)
	mux.HandleFunc("GET /sign-in/dev", bffHandler.DevSignIn)

	adminDir := adminstatic.ResolveDir(
		httpx.Env("ADMIN_STATIC_DIR", ""),
		"../node/apps/admin/dist",
		"../../node/apps/admin/dist",
	)
	if adminDir != "" {
		adminHandler := adminstatic.Handler(adminDir)
		mux.Handle("/admin/", adminHandler)
		mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/items", http.StatusFound)
		})
		log.Info("admin spa enabled", slog.String("dir", adminDir))
	} else {
		log.Warn("admin spa disabled (build: cd node && pnpm --filter @ting/admin build)")
	}

	mux.Handle("/", router)

	anonPrefixes := gwauth.AnonPrefixesFromEnv()
	log.Info("anonymous path prefixes", slog.Int("count", len(anonPrefixes)))

	h := httpx.Chain(mux,
		gwauth.Authenticate(verifier, sessions, anonPrefixes),
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8080")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}
