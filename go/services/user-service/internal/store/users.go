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

// UpdateDisplayName sets the profile display name for the authenticated user.
func (s *Users) UpdateDisplayName(ctx context.Context, id identity.Identity, displayName string) (User, error) {
	if s.pool == nil {
		return User{}, fmt.Errorf("database not connected")
	}
	if id.UserID == "" {
		return User{}, fmt.Errorf("user id required")
	}

	// Ensure row exists (first login may only PATCH from a client).
	if _, err := s.GetOrCreate(ctx, id); err != nil {
		return User{}, err
	}

	const q = `
UPDATE users
SET display_name = $3, tenant_id = $2, updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, display_name, created_at, updated_at`

	var u User
	err := s.pool.QueryRow(ctx, q, id.UserID, id.TenantID, displayName).Scan(
		&u.ID, &u.TenantID, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}

// ListByTenant returns user profiles in a tenant (newest first).
func (s *Users) ListByTenant(ctx context.Context, tenantID string, limit int) ([]User, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	const q = `
SELECT id, tenant_id, display_name, created_at, updated_at
FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2`

	rows, err := s.pool.Query(ctx, q, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	out := make([]User, 0, limit)
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
