package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ting-boundless/boundless/pkg/audit"
)

// Events persists audit events append-only.
type Events struct {
	pool *pgxpool.Pool
}

// NewEvents creates an Events store.
func NewEvents(pool *pgxpool.Pool) *Events {
	return &Events{pool: pool}
}

// Insert stores an audit event. Duplicate ids are ignored (idempotent).
func (s *Events) Insert(ctx context.Context, e audit.Event) error {
	if s.pool == nil {
		return fmt.Errorf("database not connected")
	}
	data, err := json.Marshal(e.Data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}
	if len(data) == 0 || string(data) == "null" {
		data = []byte("{}")
	}

	const q = `
INSERT INTO audit_events (id, source, type, subject, event_time, tenant_id, actor_user_id, data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO NOTHING`

	_, err = s.pool.Exec(ctx, q,
		e.ID, e.Source, e.Type, nullString(e.Subject), e.Time,
		nullString(e.TenantID), nullString(e.ActorUserID), data,
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Ping verifies database connectivity.
func (s *Events) Ping(ctx context.Context) error {
	if s.pool == nil {
		return pgx.ErrNoRows
	}
	return s.pool.Ping(ctx)
}
