package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// ConnectResult holds S3 endpoint probe state.
type ConnectResult struct {
	Endpoint string
	Probe    func(context.Context) error
}

// Connect registers an S3 readiness probe (TCP check). Cloud placeholders fail readyz.
func Connect(ctx context.Context, log *slog.Logger) ConnectResult {
	endpoint := EndpointFromEnv()
	if config.IsPlaceholder(endpoint) {
		log.Warn("object storage skipped (cloud placeholder)", slog.String("endpoint", endpoint))
		return ConnectResult{
			Endpoint: endpoint,
			Probe: func(context.Context) error {
				return fmt.Errorf("object storage not configured (placeholder endpoint %q)", endpoint)
			},
		}
	}
	if err := Ping(ctx, endpoint); err != nil {
		log.Error("object storage unreachable", slog.Any("error", err), slog.String("endpoint", endpoint))
		return ConnectResult{
			Endpoint: endpoint,
			Probe:    func(context.Context) error { return err },
		}
	}
	log.Info("object storage reachable", slog.String("endpoint", endpoint))
	return ConnectResult{
		Endpoint: endpoint,
		Probe:    func(ctx context.Context) error { return Ping(ctx, endpoint) },
	}
}

// RegisterHealth adds an object storage readiness check.
func RegisterHealth(health *httpx.Health, probe func(context.Context) error) {
	if probe == nil {
		return
	}
	health.Register(httpx.Check{Name: "s3", Probe: probe})
}
