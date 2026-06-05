package cache

import (
	"context"
	"fmt"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// ConnectResult holds a Redis client or a placeholder readiness probe.
type ConnectResult struct {
	Client *goredis.Client
	Probe  func(context.Context) error
}

// Connect opens Redis for local dev, or registers a failing probe for cloud placeholders.
func Connect(ctx context.Context, log *slog.Logger) ConnectResult {
	addr := httpx.Env("REDIS_ADDR", "localhost:6379")
	if config.IsPlaceholder(addr) {
		log.Warn("redis skipped (cloud placeholder)", slog.String("addr", addr))
		return ConnectResult{
			Probe: func(context.Context) error {
				return fmt.Errorf("redis not configured (placeholder addr %q)", addr)
			},
		}
	}

	client, err := OpenFromEnv(ctx)
	if err != nil {
		log.Error("redis connection failed", slog.Any("error", err), slog.String("addr", addr))
		return ConnectResult{
			Probe: func(context.Context) error { return err },
		}
	}
	log.Info("redis connected", slog.String("addr", addr))
	return ConnectResult{
		Client: client,
		Probe:  func(ctx context.Context) error { return Ping(ctx, client) },
	}
}

// RegisterHealth adds a redis readiness check.
func RegisterHealth(health *httpx.Health, probe func(context.Context) error) {
	if probe == nil {
		return
	}
	health.Register(httpx.Check{Name: "redis", Probe: probe})
}
