package cache

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// OpenFromEnv connects to Redis using REDIS_ADDR (host:port).
func OpenFromEnv(ctx context.Context) (*goredis.Client, error) {
	addr := httpx.Env("REDIS_ADDR", "localhost:6379")
	if config.IsPlaceholder(addr) {
		return nil, fmt.Errorf("redis addr is placeholder (%q)", addr)
	}
	client := goredis.NewClient(&goredis.Options{Addr: addr})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return client, nil
}

// Ping checks Redis connectivity.
func Ping(ctx context.Context, client *goredis.Client) error {
	return client.Ping(ctx).Err()
}
