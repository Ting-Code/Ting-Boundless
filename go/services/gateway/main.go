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

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/cache"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/oidc"
	"github.com/ting-boundless/boundless/pkg/revocation"
	"github.com/ting-boundless/boundless/services/gateway/internal/adminstatic"
	gwauth "github.com/ting-boundless/boundless/services/gateway/internal/auth"
	"github.com/ting-boundless/boundless/services/gateway/internal/bff"
	"github.com/ting-boundless/boundless/services/gateway/internal/identityresolve"
	"github.com/ting-boundless/boundless/services/gateway/internal/proxy"
	"github.com/ting-boundless/boundless/services/gateway/internal/ratelimit"
	"github.com/ting-boundless/boundless/services/gateway/internal/session"
	"github.com/ting-boundless/boundless/services/gateway/internal/siteproxy"
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
	revocations := revocation.NewStore(rd.Client)
	bffHandler := bff.New(oidcCfg, sessions, verifier, authCfg, revocations, log)

	if oidcCfg.Ready() {
		log.Info("oidc bff enabled", slog.String("redirect_uri", oidcCfg.RedirectURI))
	} else {
		log.Warn("oidc bff disabled (set OIDC_CLIENT_ID and OIDC_CLIENT_SECRET for Logto)")
	}

	health := httpx.NewHealth()
	cache.RegisterHealth(health, rd.Probe)

	internalToken, ok := httpx.LoadInternalToken(log)
	if !ok {
		return
	}

	routes := proxy.Routes{
		"/v1/users/":           httpx.Env("USER_SERVICE_URL", "http://127.0.0.1:8081"),
		"/v1/business/":        httpx.Env("BUSINESS_SERVICE_URL", "http://127.0.0.1:3005"),
		"/v1/files/":           httpx.Env("FILE_SERVICE_URL", "http://127.0.0.1:8083"),
		"/v1/auth/":            httpx.Env("AUTH_SERVICE_URL", "http://127.0.0.1:8084"),
		"/v1/audit/":           httpx.Env("AUDIT_SERVICE_URL", "http://127.0.0.1:8085"),
		"/internal/webhooks/":  httpx.Env("AUTH_SERVICE_URL", "http://127.0.0.1:8084"),
	}
	log.Info("upstream routes",
		slog.String("users", routes["/v1/users/"]),
		slog.String("business", routes["/v1/business/"]),
		slog.String("files", routes["/v1/files/"]),
		slog.String("auth", routes["/v1/auth/"]),
		slog.String("audit", routes["/v1/audit/"]),
	)

	router, err := proxy.New(routes, log, internalToken)
	if err != nil {
		log.Error("failed to build proxy", slog.Any("error", err))
		return
	}

	siteHandler, err := siteproxy.FromEnv(log)
	if err != nil {
		log.Error("site proxy init failed", slog.Any("error", err))
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

	mux.Handle("/", siteproxy.ComposeAPIAndSite(router, siteHandler))

	anonPrefixes := gwauth.ResolveAnonPrefixes()
	log.Info("anonymous path rules",
		slog.Int("exact", len(anonPrefixes.ExactPaths())),
		slog.Int("prefix", len(anonPrefixes.PrefixPaths())),
	)

	sensitivePrefixes := gwauth.ResolveSensitivePrefixes()
	log.Info("sensitive path rules", slog.Int("prefix", len(sensitivePrefixes.PrefixPaths())))

	identityResolver := identityresolve.FromEnv()

	entryAudit := audit.NewAsync(audit.NewHTTPEmitter(audit.HTTPEmitterConfig{
		BaseURL: httpx.Env("AUDIT_SERVICE_URL", "http://127.0.0.1:8085"),
		Token:   internalToken,
	}), log)
	if entryAudit == nil {
		log.Warn("AUDIT_SERVICE_URL not set; gateway entry audit events disabled")
	}

	rateCfg := ratelimit.ConfigFromEnv()
	log.Info("rate limits",
		slog.Bool("enabled", rateCfg.Enabled),
		slog.Float64("auth_rps", rateCfg.AuthRPS),
		slog.Float64("general_rps", rateCfg.GeneralRPS),
		slog.Bool("redis", rd.Client != nil && httpx.Env("GATEWAY_RATE_LIMIT_REDIS", "true") != "false"),
	)

	h := httpx.Chain(mux,
		ratelimit.Middleware(rateCfg, entryAudit, ratelimit.NewLimiter(rateCfg, rd.Client, log)),
		gwauth.Authenticate(verifier, sessions, anonPrefixes, identityResolver, entryAudit, revocations, sensitivePrefixes),
		httpx.Recover(log),
		httpx.AccessLog(log),
		httpx.TraceContext,
	)

	addr := httpx.Env("HTTP_ADDR", ":8080")
	if err := httpx.RunService(addr, serviceName, h, log); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}
