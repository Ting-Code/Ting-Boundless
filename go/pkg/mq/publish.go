package mq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher publishes jobs to the platform work queue.
type Publisher struct {
	ch    *amqp.Channel
	queue WorkQueue
}

// NewPublisher opens a channel and declares topology.
func NewPublisher(conn *amqp.Connection, queue WorkQueue) (*Publisher, error) {
	if conn == nil {
		return nil, fmt.Errorf("rabbitmq connection is nil")
	}
	if queue.Exchange == "" {
		queue = DefaultWorkQueue()
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := queue.Declare(ch); err != nil {
		_ = ch.Close()
		return nil, err
	}
	return &Publisher{ch: ch, queue: queue}, nil
}

// Publish sends a job to the work exchange.
func (p *Publisher) Publish(ctx context.Context, job Job) error {
	if p == nil || p.ch == nil {
		return fmt.Errorf("publisher unavailable")
	}
	body, err := job.Marshal()
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, p.queue.Exchange, p.queue.RoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

// Close closes the publisher channel.
func (p *Publisher) Close() error {
	if p == nil || p.ch == nil {
		return nil
	}
	return p.ch.Close()
}
