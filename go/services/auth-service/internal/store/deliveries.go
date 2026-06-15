package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Deliveries tracks processed webhook deliveries for idempotency.
type Deliveries struct {
	pool *pgxpool.Pool
}

// NewDeliveries wraps a Postgres pool.
func NewDeliveries(pool *pgxpool.Pool) *Deliveries {
	return &Deliveries{pool: pool}
}

// TryRecord inserts a delivery key. Returns false if already processed.
func (d *Deliveries) TryRecord(ctx context.Context, deliveryKey string) (bool, error) {
	if d == nil || d.pool == nil {
		return false, fmt.Errorf("deliveries store not configured")
	}
	tag, err := d.pool.Exec(ctx,
		`INSERT INTO webhook_deliveries (delivery_key) VALUES ($1) ON CONFLICT (delivery_key) DO NOTHING`,
		deliveryKey,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// Ping verifies database connectivity.
func (d *Deliveries) Ping(ctx context.Context) error {
	if d.pool == nil {
		return pgx.ErrNoRows
	}
	return d.pool.Ping(ctx)
}
