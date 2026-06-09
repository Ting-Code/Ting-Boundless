package mq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// OpenFromEnv dials RabbitMQ using RABBITMQ_URL.
func OpenFromEnv() (*amqp.Connection, error) {
	url := httpx.Env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	if config.IsPlaceholder(url) {
		return nil, fmt.Errorf("rabbitmq url is placeholder (%q)", url)
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	return conn, nil
}

// Ping verifies RabbitMQ is reachable within timeout.
func Ping(ctx context.Context, url string) error {
	if config.IsPlaceholder(url) {
		return fmt.Errorf("rabbitmq url is placeholder (%q)", url)
	}
	if url == "" {
		url = httpx.Env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	}
	done := make(chan error, 1)
	go func() {
		conn, err := amqp.Dial(url)
		if err != nil {
			done <- err
			return
		}
		_ = conn.Close()
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("rabbitmq ping: %w", err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		return fmt.Errorf("rabbitmq ping: timeout")
	}
}
