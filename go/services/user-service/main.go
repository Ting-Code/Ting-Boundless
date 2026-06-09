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
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
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

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("GET /v1/users/me", identity.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleMe(w, r, users)
	})))

	internalToken := httpx.Env("INTERNAL_API_TOKEN", "")

	h := httpx.Chain(mux,
		httpx.GatewayTrust(internalToken),
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8081")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleMe(w http.ResponseWriter, r *http.Request, users *store.Users) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		errs.Write(w, id.RequestID, errs.Unauthorized("unauthenticated", "authentication required"))
		return
	}
	if users == nil {
		errs.Write(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}

	u, err := users.GetOrCreate(r.Context(), id)
	if err != nil {
		errs.Write(w, id.RequestID, errs.Internal("user_lookup_failed", "failed to load user profile"))
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"user_id":      u.ID,
		"tenant_id":    u.TenantID,
		"display_name": u.DisplayName,
		"roles":        id.Roles,
		"created_at":   u.CreatedAt,
	})
}
