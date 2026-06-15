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
	"strconv"
	"time"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/cache"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/revocation"
	"github.com/ting-boundless/boundless/services/auth-service/internal/identityapi"
	"github.com/ting-boundless/boundless/services/auth-service/internal/jwks"
	"github.com/ting-boundless/boundless/services/auth-service/internal/logto"
	"github.com/ting-boundless/boundless/services/auth-service/internal/miniprogram"
	"github.com/ting-boundless/boundless/services/auth-service/internal/store"
	"github.com/ting-boundless/boundless/services/auth-service/internal/wechat"
)

const serviceName = "auth-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	pgCfg := db.ConfigFromEnv("")
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
		if err := db.RunMigrations(pgCfg, serviceName); err != nil {
			log.Error("migrations failed", slog.Any("error", err))
			return
		}
	}
	rd := cache.Connect(ctx, log)
	if rd.Client != nil {
		defer rd.Client.Close()
	}

	issuer, err := auth.NewIssuer(auth.IssuerConfig{
		Issuer:          httpx.Env("AUTH_OIDC_ISSUER", "http://127.0.0.1:8084/oidc"),
		Audience:        httpx.Env("OIDC_AUDIENCE", ""),
		PrivateKeyPEM:   httpx.Env("AUTH_JWT_PRIVATE_KEY_PEM", ""),
		GenerateIfEmpty: httpx.Env("AUTH_JWT_PRIVATE_KEY_PEM", "") == "",
	})
	if err != nil {
		log.Error("jwt issuer init failed", slog.Any("error", err))
		return
	}
	if httpx.Env("AUTH_JWT_PRIVATE_KEY_PEM", "") == "" {
		log.Warn("AUTH_JWT_PRIVATE_KEY_PEM not set; using ephemeral dev RSA key (restart invalidates tokens)")
	}

	wx := wechat.NewClient(wechat.Config{
		AppID:     httpx.Env("WECHAT_APP_ID", ""),
		AppSecret: httpx.Env("WECHAT_APP_SECRET", ""),
		MockMode:  httpx.Env("WECHAT_MOCK_MODE", "false") == "true",
	})

	var identities *store.IdentityStore
	var deliveries *store.Deliveries
	if pg.DB != nil {
		pool := pg.DB.Pool()
		identities = store.NewIdentityStore(pool)
		deliveries = store.NewDeliveries(pool)
	}

	internalToken, ok := httpx.LoadInternalToken(log)
	if !ok {
		return
	}
	auditEmitter := audit.NewHTTPEmitter(audit.HTTPEmitterConfig{
		BaseURL: httpx.Env("AUDIT_SERVICE_URL", "http://127.0.0.1:8085"),
		Token:   internalToken,
	})
	if !auditEmitter.Enabled() {
		log.Warn("AUDIT_SERVICE_URL not set; identity audit events will not be delivered")
	}

	logtoSigningKey := httpx.Env("LOGTO_WEBHOOK_SIGNING_KEY", "")
	skipVerify := httpx.Env("LOGTO_WEBHOOK_SKIP_VERIFY", "false") == "true"
	if logtoSigningKey == "" && skipVerify {
		log.Warn("LOGTO_WEBHOOK_SKIP_VERIFY=true without signing key; webhook signatures not checked")
	}

	logtoHook := logto.NewHandler(logto.Config{
		SigningKey: logtoSigningKey,
		SkipVerify: skipVerify,
		Identities: identities,
		Deliveries: deliveries,
		Audit:      auditEmitter,
		Revocations: revocation.NewStore(rd.Client),
		Log:        log,
	})

	ttl := time.Hour
	if s := httpx.Env("AUTH_JWT_ACCESS_TTL_SECONDS", ""); s != "" {
		if sec, err := strconv.Atoi(s); err == nil && sec > 0 {
			ttl = time.Duration(sec) * time.Second
		}
	}

	mp := miniprogram.NewHandler(miniprogram.Config{
		WeChat:   wx,
		Identity: identities,
		Issuer:   issuer,
		Audit:    auditEmitter,
		TTL:      ttl,
		Log:      log,
	})

	health := httpx.NewHealth()
	db.RegisterHealth(health, "postgres", pg.Probe)
	cache.RegisterHealth(health, rd.Probe)

	mux := http.NewServeMux()
	health.Handler(mux)

	mux.Handle("POST /internal/webhooks/logto",
		httpx.InternalAuth(internalToken)(logtoHook),
	)
	mux.Handle("POST /internal/identity/resolve",
		httpx.InternalAuth(internalToken)(identityapi.NewResolveHandler(identities)),
	)
	mux.HandleFunc("POST /v1/auth/miniprogram/login", mp.Login)
	mux.Handle("GET /v1/auth/jwks", jwks.NewHandler(issuer))

	h := httpx.Chain(mux,
		httpx.GatewayTrust(internalToken),
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
		httpx.TraceContext,
	)

	addr := httpx.Env("HTTP_ADDR", ":8084")
	if err := httpx.RunService(addr, serviceName, h, log); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}
