package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ting-boundless/boundless/pkg/identity"
)

// User is a persisted user profile row.
type User struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Users provides data access for the user domain.
type Users struct {
	pool *pgxpool.Pool
}

// NewUsers creates a Users store.
func NewUsers(pool *pgxpool.Pool) *Users {
	return &Users{pool: pool}
}

// GetOrCreate returns the user row, creating a minimal profile on first access.
func (s *Users) GetOrCreate(ctx context.Context, id identity.Identity) (User, error) {
	if s.pool == nil {
		return User{}, fmt.Errorf("database not connected")
	}
	if id.UserID == "" {
		return User{}, fmt.Errorf("user id required")
	}

	const q = `
INSERT INTO users (id, tenant_id, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET
    tenant_id = EXCLUDED.tenant_id,
    updated_at = now()
RETURNING id, tenant_id, display_name, created_at, updated_at`

	var u User
	err := s.pool.QueryRow(ctx, q, id.UserID, id.TenantID, id.UserID).Scan(
		&u.ID, &u.TenantID, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("get or create user: %w", err)
	}
	return u, nil
}
