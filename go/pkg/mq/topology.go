package mq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// WorkQueue declares a durable work queue with a dead-letter queue (DLQ).
type WorkQueue struct {
	Exchange   string
	Queue      string
	DLX        string
	DLQ        string
	RoutingKey string
}

// DefaultWorkQueue returns the platform jobs queue topology.
func DefaultWorkQueue() WorkQueue {
	prefix := httpx.Env("RABBITMQ_QUEUE_PREFIX", "ting.jobs")
	return WorkQueue{
		Exchange:   prefix,
		Queue:      prefix + ".platform",
		DLX:        prefix + ".dlx",
		DLQ:        prefix + ".platform.dlq",
		RoutingKey: "platform",
	}
}

// Declare ensures exchange, DLQ, and main queue exist on the channel.
func (q WorkQueue) Declare(ch *amqp.Channel) error {
	if ch == nil {
		return fmt.Errorf("rabbitmq channel is nil")
	}
	if err := ch.ExchangeDeclare(q.Exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %q: %w", q.Exchange, err)
	}
	if err := ch.ExchangeDeclare(q.DLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx %q: %w", q.DLX, err)
	}
	if _, err := ch.QueueDeclare(q.DLQ, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlq %q: %w", q.DLQ, err)
	}
	if err := ch.QueueBind(q.DLQ, q.RoutingKey, q.DLX, false, nil); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}
	args := amqp.Table{
		"x-dead-letter-exchange":    q.DLX,
		"x-dead-letter-routing-key": q.RoutingKey,
	}
	if _, err := ch.QueueDeclare(q.Queue, true, false, false, false, args); err != nil {
		return fmt.Errorf("declare queue %q: %w", q.Queue, err)
	}
	if err := ch.QueueBind(q.Queue, q.RoutingKey, q.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}
	return nil
}
