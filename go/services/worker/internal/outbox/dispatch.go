// Package outbox dispatches business_outbox rows to audit-service.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ting-boundless/boundless/pkg/audit"
)

const businessSource = "business-service"

// Dispatcher polls unpublished outbox rows and emits audit events.
type Dispatcher struct {
	pool   *pgxpool.Pool
	audit  audit.Emitter
	log    *slog.Logger
	period time.Duration
}

// Config configures the outbox dispatcher.
type Config struct {
	Pool   *pgxpool.Pool
	Audit  audit.Emitter
	Log    *slog.Logger
	Period time.Duration
}

// NewDispatcher builds a dispatcher. Returns nil when pool or audit is unavailable.
func NewDispatcher(cfg Config) *Dispatcher {
	if cfg.Pool == nil || cfg.Audit == nil {
		return nil
	}
	if http, ok := cfg.Audit.(*audit.HTTPEmitter); ok && !http.Enabled() {
		return nil
	}
	period := cfg.Period
	if period <= 0 {
		period = 2 * time.Second
	}
	return &Dispatcher{
		pool:   cfg.Pool,
		audit:  cfg.Audit,
		log:    cfg.Log,
		period: period,
	}
}

// Run polls until ctx is cancelled.
func (d *Dispatcher) Run(ctx context.Context) {
	if d == nil {
		return
	}
	ticker := time.NewTicker(d.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.dispatchOnce(ctx); err != nil && d.log != nil {
				d.log.Warn("outbox dispatch failed", slog.Any("error", err))
			}
		}
	}
}

func (d *Dispatcher) dispatchOnce(ctx context.Context) error {
	rows, err := d.pool.Query(ctx, `
SELECT id::text, event_type, payload, created_at
FROM business_outbox
WHERE published_at IS NULL
ORDER BY created_at
LIMIT 50`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var firstErr error
	for rows.Next() {
		var id, eventType string
		var payload []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &eventType, &payload, &createdAt); err != nil {
			return err
		}
		if err := d.publishRow(ctx, id, eventType, payload, createdAt); err != nil {
			if d.log != nil {
				d.log.Warn("outbox row failed", slog.String("id", id), slog.Any("error", err))
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if d.log != nil {
			d.log.Info("outbox delivered", slog.String("id", id), slog.String("type", eventType))
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return firstErr
}

func (d *Dispatcher) publishRow(ctx context.Context, id, eventType string, payload []byte, createdAt time.Time) error {
	ev, err := rowToEvent(id, eventType, payload, createdAt)
	if err != nil {
		return err
	}
	if err := d.audit.Emit(ctx, ev); err != nil {
		return err
	}
	_, err = d.pool.Exec(ctx,
		`UPDATE business_outbox SET published_at = now() WHERE id = $1::uuid AND published_at IS NULL`,
		id,
	)
	return err
}

func rowToEvent(id, eventType string, payload []byte, createdAt time.Time) (audit.Event, error) {
	var data map[string]any
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &data); err != nil {
			return audit.Event{}, fmt.Errorf("payload json: %w", err)
		}
	}
	ev := audit.Event{
		ID:     id,
		Source: businessSource,
		Type:   eventType,
		Time:   createdAt.UTC(),
		Data:   data,
	}
	if ev.Data == nil {
		ev.Data = map[string]any{}
	}
	if v, _ := data["actor_user_id"].(string); v != "" {
		ev.ActorUserID = v
	}
	if v, _ := data["tenant_id"].(string); v != "" {
		ev.TenantID = v
	}
	if v, _ := data["item_id"].(string); v != "" {
		ev.Subject = "item:" + v
	}
	return ev, nil
}
