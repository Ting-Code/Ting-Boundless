package mq

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// ConnectResult holds a RabbitMQ connection or a placeholder readiness probe.
type ConnectResult struct {
	Conn  *amqp.Connection
	Probe func(context.Context) error
}

// Connect dials RabbitMQ for local dev, or registers a failing probe for cloud placeholders.
func Connect(ctx context.Context, log *slog.Logger) ConnectResult {
	url := httpx.Env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	if config.IsPlaceholder(url) {
		log.Warn("rabbitmq skipped (cloud placeholder)", slog.String("url", redactURL(url)))
		return ConnectResult{
			Probe: func(context.Context) error {
				return fmt.Errorf("rabbitmq not configured (placeholder url)")
			},
		}
	}

	conn, err := OpenFromEnv()
	if err != nil {
		log.Error("rabbitmq connection failed", slog.Any("error", err))
		return ConnectResult{
			Probe: func(context.Context) error { return err },
		}
	}
	log.Info("rabbitmq connected")
	return ConnectResult{
		Conn: conn,
		Probe: func(ctx context.Context) error {
			return Ping(ctx, url)
		},
	}
}

// RegisterHealth adds a rabbitmq readiness check.
func RegisterHealth(health *httpx.Health, probe func(context.Context) error) {
	if probe == nil {
		return
	}
	health.Register(httpx.Check{Name: "rabbitmq", Probe: probe})
}

func redactURL(raw string) string {
	if config.IsPlaceholder(raw) {
		return raw
	}
	return "amqp://***"
}
