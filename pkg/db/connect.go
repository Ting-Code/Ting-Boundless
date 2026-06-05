package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// ConnectResult holds an opened pool or a placeholder readiness probe.
type ConnectResult struct {
	DB    *Postgres
	Probe func(context.Context) error
}

// Connect opens PostgreSQL when configured for local use, or registers a failing
// readiness probe when the host is a cloud placeholder.
func Connect(ctx context.Context, log *slog.Logger, database string) ConnectResult {
	cfg := ConfigFromEnv(database)
	if config.IsPlaceholder(cfg.Host) {
		log.Warn("postgres skipped (cloud placeholder)",
			slog.String("host", cfg.Host),
			slog.String("database", cfg.Database),
		)
		return ConnectResult{
			Probe: func(context.Context) error {
				return fmt.Errorf("postgres not configured (placeholder host %q)", cfg.Host)
			},
		}
	}

	pg, err := Open(ctx, cfg)
	if err != nil {
		log.Error("postgres connection failed",
			slog.Any("error", err),
			slog.String("host", cfg.Host),
			slog.String("database", cfg.Database),
		)
		return ConnectResult{
			Probe: func(context.Context) error { return err },
		}
	}
	log.Info("postgres connected",
		slog.String("host", cfg.Host),
		slog.String("database", cfg.Database),
	)
	return ConnectResult{
		DB:    pg,
		Probe: pg.Ping,
	}
}

// RegisterHealth adds a postgres readiness check when probe is set.
func RegisterHealth(health *httpx.Health, name string, probe func(context.Context) error) {
	if probe == nil {
		return
	}
	health.Register(httpx.Check{Name: name, Probe: probe})
}
