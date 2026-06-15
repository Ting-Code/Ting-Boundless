package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ting-boundless/boundless/pkg/audit"
)

// EventRow is a persisted audit event for read APIs.
type EventRow struct {
	ID          string         `json:"id"`
	Source      string         `json:"source"`
	Type        string         `json:"type"`
	Subject     string         `json:"subject,omitempty"`
	Time        time.Time      `json:"time"`
	TenantID    string         `json:"tenant_id,omitempty"`
	ActorUserID string         `json:"actor_user_id,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
	ReceivedAt  time.Time      `json:"received_at"`
}

// ListFilter scopes a read query.
type ListFilter struct {
	TenantID string
	Type     string
	Source   string
	Limit    int
}

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

// List returns recent audit events matching the filter (newest first).
func (s *Events) List(ctx context.Context, f ListFilter) ([]EventRow, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("database not connected")
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	var b strings.Builder
	b.WriteString(`
SELECT id, source, type, COALESCE(subject, ''), event_time,
       COALESCE(tenant_id, ''), COALESCE(actor_user_id, ''), data, received_at
FROM audit_events
WHERE 1=1`)

	args := make([]any, 0, 4)
	n := 1
	if f.TenantID != "" {
		fmt.Fprintf(&b, " AND (tenant_id = $%d OR tenant_id IS NULL OR tenant_id = '')", n)
		args = append(args, f.TenantID)
		n++
	}
	if f.Type != "" {
		fmt.Fprintf(&b, " AND type = $%d", n)
		args = append(args, f.Type)
		n++
	}
	if f.Source != "" {
		fmt.Fprintf(&b, " AND source = $%d", n)
		args = append(args, f.Source)
		n++
	}
	fmt.Fprintf(&b, " ORDER BY event_time DESC, id DESC LIMIT $%d", n)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	out := make([]EventRow, 0, limit)
	for rows.Next() {
		var row EventRow
		var data []byte
		if err := rows.Scan(
			&row.ID, &row.Source, &row.Type, &row.Subject, &row.Time,
			&row.TenantID, &row.ActorUserID, &data, &row.ReceivedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		if len(data) > 0 && string(data) != "null" {
			if err := json.Unmarshal(data, &row.Data); err != nil {
				return nil, fmt.Errorf("unmarshal data: %w", err)
			}
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
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
