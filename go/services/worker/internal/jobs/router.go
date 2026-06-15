package jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ting-boundless/boundless/pkg/mq"
)

type effectStore interface {
	Record(ctx context.Context, job mq.Job, resourceID string) (bool, error)
}

// Router dispatches platform async jobs to typed handlers.
type Router struct {
	log     *slog.Logger
	effects effectStore
}

// NewRouter builds the default job handler.
func NewRouter(log *slog.Logger, effects effectStore) mq.Handler {
	return (&Router{log: log, effects: effects}).Handle
}

func (r *Router) Handle(ctx context.Context, job mq.Job) error {
	switch job.Type {
	case "ping":
		if r.log != nil {
			r.log.Info("ping job", slog.String("id", job.ID))
		}
		return nil
	case "business.item.created", "business.item.updated", "business.item.deleted":
		return r.handleBusinessItem(ctx, job)
	default:
		if r.log != nil {
			r.log.Warn("unknown job type (acked)",
				slog.String("id", job.ID),
				slog.String("type", job.Type),
			)
		}
		return nil
	}
}

func (r *Router) handleBusinessItem(ctx context.Context, job mq.Job) error {
	itemID, err := requireString(job.Payload, "item_id")
	if err != nil {
		return err
	}

	if r.effects == nil {
		if r.log != nil {
			r.log.Warn("job effects store unavailable; skipping persist",
				slog.String("id", job.ID),
				slog.String("type", job.Type),
			)
		}
		return nil
	}

	inserted, err := r.effects.Record(ctx, job, itemID)
	if err != nil {
		return fmt.Errorf("record business item effect: %w", err)
	}

	if r.log != nil {
		attrs := []any{
			slog.String("id", job.ID),
			slog.String("type", job.Type),
			slog.String("item_id", itemID),
			slog.String("tenant_id", job.Tenant),
			slog.String("actor_user_id", job.Actor),
			slog.Bool("first_delivery", inserted),
		}
		if title, ok := job.Payload["title"].(string); ok && title != "" {
			attrs = append(attrs, slog.String("title", title))
		}
		r.log.Info("business item side-effect", attrs...)
	}
	return nil
}

func requireString(payload map[string]any, key string) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("job payload missing %s", key)
	}
	raw, ok := payload[key]
	if !ok {
		return "", fmt.Errorf("job payload missing %s", key)
	}
	s, ok := raw.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("job payload %s must be a non-empty string", key)
	}
	return s, nil
}
