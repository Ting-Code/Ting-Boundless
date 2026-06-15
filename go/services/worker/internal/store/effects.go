package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ting-boundless/boundless/pkg/mq"
)

// Effect is a persisted async job side-effect row.
type Effect struct {
	JobID       string
	JobType     string
	TenantID    string
	ActorUserID string
	ResourceID  string
}

// Effects records idempotent platform job side-effects.
type Effects struct {
	pool *pgxpool.Pool
}

// NewEffects creates an Effects store.
func NewEffects(pool *pgxpool.Pool) *Effects {
	return &Effects{pool: pool}
}

// Record inserts a job effect once. Returns true when a new row was written.
func (s *Effects) Record(ctx context.Context, job mq.Job, resourceID string) (bool, error) {
	if s.pool == nil {
		return false, fmt.Errorf("database not connected")
	}
	if job.ID == "" || job.Type == "" {
		return false, fmt.Errorf("job id and type are required")
	}

	payload, err := json.Marshal(job.Payload)
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}
	if len(payload) == 0 {
		payload = []byte("{}")
	}

	const q = `
INSERT INTO worker_job_effects (job_id, job_type, tenant_id, actor_user_id, resource_id, payload)
VALUES ($1, $2, $3, $4, $5, $6::jsonb)
ON CONFLICT (job_id) DO NOTHING
RETURNING job_id`

	var inserted string
	err = s.pool.QueryRow(ctx, q,
		job.ID, job.Type, job.Tenant, job.Actor, resourceID, payload,
	).Scan(&inserted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("record job effect: %w", err)
	}
	return true, nil
}

// EffectRow is a persisted job effect for read APIs.
type EffectRow struct {
	JobID       string         `json:"job_id"`
	JobType     string         `json:"job_type"`
	TenantID    string         `json:"tenant_id"`
	ActorUserID string         `json:"actor_user_id"`
	ResourceID  string         `json:"resource_id"`
	Payload     map[string]any `json:"payload,omitempty"`
	ProcessedAt time.Time      `json:"processed_at"`
}

// ListFilter scopes a job-effects query.
type ListFilter struct {
	TenantID string
	JobType  string
	Limit    int
}

// List returns recent job effects (newest first).
func (s *Effects) List(ctx context.Context, f ListFilter) ([]EffectRow, error) {
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
SELECT job_id, job_type, COALESCE(tenant_id, ''), COALESCE(actor_user_id, ''),
       COALESCE(resource_id, ''), payload, processed_at
FROM worker_job_effects
WHERE 1=1`)

	args := make([]any, 0, 3)
	n := 1
	if f.TenantID != "" {
		fmt.Fprintf(&b, " AND tenant_id = $%d", n)
		args = append(args, f.TenantID)
		n++
	}
	if f.JobType != "" {
		fmt.Fprintf(&b, " AND job_type = $%d", n)
		args = append(args, f.JobType)
		n++
	}
	fmt.Fprintf(&b, " ORDER BY processed_at DESC LIMIT $%d", n)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list job effects: %w", err)
	}
	defer rows.Close()

	out := make([]EffectRow, 0, limit)
	for rows.Next() {
		var row EffectRow
		var payload []byte
		if err := rows.Scan(
			&row.JobID, &row.JobType, &row.TenantID, &row.ActorUserID,
			&row.ResourceID, &payload, &row.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("scan job effect: %w", err)
		}
		if len(payload) > 0 && string(payload) != "null" {
			if err := json.Unmarshal(payload, &row.Payload); err != nil {
				return nil, fmt.Errorf("unmarshal payload: %w", err)
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
