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
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/services/gateway/internal/proxy"
)

const serviceName = "gateway"

func main() {
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	health := httpx.NewHealth()
	// Intentionally NOT registering Logto JWKS as a readiness check: JWKS is
	// cached, so a Logto outage degrades login only, it must not fail readiness.

	routes := proxy.Routes{
		"/v1/users/":    httpx.Env("USER_SERVICE_URL", "http://user-service:8080"),
		"/v1/business/": httpx.Env("BUSINESS_SERVICE_URL", "http://business-service:8080"),
		"/v1/files/":    httpx.Env("FILE_SERVICE_URL", "http://file-service:8080"),
	}

	router, err := proxy.New(routes, log)
	if err != nil {
		log.Error("failed to build proxy", slog.Any("error", err))
		return
	}

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("/", router)

	// Order matters: strip+verify+inject identity BEFORE proxying downstream.
	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
		authenticate(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8080")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

// authenticate strips untrusted identity headers, verifies the caller's
// credential, and injects the trusted identity context downstream.
//
// TODO: implement real verification:
//   - Bearer JWT: validate via cached JWKS (OIDC_JWKS_URL), check issuer/audience
//   - Web cookie: validate BFF session
//   - check Redis revocation list for sensitive paths
func authenticate(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Never trust client-supplied identity headers.
			identity.StripUntrusted(r.Header)

			// 2. Verify credential -> resolve identity (placeholder).
			id := identity.Identity{RequestID: r.Header.Get(identity.HeaderRequestID)}
			// id = verifyBearerOrCookie(r) ...

			// 3. Re-inject the request id and the (verified) identity.
			id.Inject(r.Header)

			next.ServeHTTP(w, r)
		})
	}
}
