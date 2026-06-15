package mq

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// ConsumerConfig wires a RabbitMQ work-queue consumer.
type ConsumerConfig struct {
	Conn       *amqp.Connection
	Queue      WorkQueue
	Handler    Handler
	Log        *slog.Logger
	Prefetch   int
	MaxRetries int
}

// Consumer runs a single-queue AMQP consumer with DLQ on exhaustion.
type Consumer struct {
	cfg ConsumerConfig
}

// NewConsumer builds a consumer. Returns nil when conn or handler is missing.
func NewConsumer(cfg ConsumerConfig) *Consumer {
	if cfg.Conn == nil || cfg.Handler == nil {
		return nil
	}
	if cfg.Queue.Exchange == "" {
		cfg.Queue = DefaultWorkQueue()
	}
	if cfg.Prefetch <= 0 {
		cfg.Prefetch = envInt("RABBITMQ_PREFETCH", 10)
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = envInt("RABBITMQ_MAX_RETRIES", 3)
	}
	return &Consumer{cfg: cfg}
}

// Run blocks until ctx is cancelled or the AMQP channel closes.
func (c *Consumer) Run(ctx context.Context) error {
	if c == nil {
		return nil
	}

	ch, err := c.cfg.Conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := c.cfg.Queue.Declare(ch); err != nil {
		return err
	}
	if err := ch.Qos(c.cfg.Prefetch, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	deliveries, err := ch.Consume(c.cfg.Queue.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %q: %w", c.cfg.Queue.Queue, err)
	}

	if c.cfg.Log != nil {
		c.cfg.Log.Info("rabbitmq consumer started",
			slog.String("queue", c.cfg.Queue.Queue),
			slog.String("dlq", c.cfg.Queue.DLQ),
			slog.Int("prefetch", c.cfg.Prefetch),
		)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbitmq deliveries channel closed")
			}
			c.handleDelivery(ctx, d)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery) {
	job, err := ParseJob(d.Body)
	if err != nil {
		if c.cfg.Log != nil {
			c.cfg.Log.Warn("invalid job payload", slog.Any("error", err))
		}
		_ = d.Nack(false, false)
		return
	}

	if err := c.cfg.Handler(ctx, job); err != nil {
		retries := deathCount(d.Headers) + 1
		requeue := retries < c.cfg.MaxRetries
		if c.cfg.Log != nil {
			c.cfg.Log.Warn("job failed",
				slog.String("id", job.ID),
				slog.String("type", job.Type),
				slog.Int("attempt", retries),
				slog.Bool("requeue", requeue),
				slog.Any("error", err),
			)
		}
		_ = d.Nack(false, requeue)
		return
	}

	if c.cfg.Log != nil {
		c.cfg.Log.Info("job completed", slog.String("id", job.ID), slog.String("type", job.Type))
	}
	_ = d.Ack(false)
}

func deathCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	raw, ok := headers["x-death"]
	if !ok {
		return 0
	}
	list, ok := raw.([]any)
	if !ok {
		return 0
	}
	total := 0
	for _, item := range list {
		m, ok := item.(amqp.Table)
		if !ok {
			continue
		}
		switch v := m["count"].(type) {
		case int64:
			total += int(v)
		case int32:
			total += int(v)
		case int:
			total += v
		case float64:
			total += int(v)
		}
	}
	return total
}

func envInt(key string, def int) int {
	v := httpx.Env(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}
