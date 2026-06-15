package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// File is persisted object metadata.
type File struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	OwnerID     string    `json:"owner_id"`
	Bucket      string    `json:"bucket"`
	ObjectKey   string    `json:"object_key"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

// Files provides metadata access for uploaded objects.
type Files struct {
	pool *pgxpool.Pool
}

// NewFiles creates a Files store.
func NewFiles(pool *pgxpool.Pool) *Files {
	return &Files{pool: pool}
}

// Insert records file metadata after a successful object upload.
func (s *Files) Insert(ctx context.Context, f File) (File, error) {
	if s.pool == nil {
		return File{}, fmt.Errorf("database not connected")
	}
	const q = `
INSERT INTO files (id, tenant_id, owner_id, bucket, object_key, content_type, size_bytes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, owner_id, bucket, object_key, content_type, size_bytes, created_at`

	var out File
	err := s.pool.QueryRow(ctx, q,
		f.ID, f.TenantID, f.OwnerID, f.Bucket, f.ObjectKey, f.ContentType, f.SizeBytes,
	).Scan(
		&out.ID, &out.TenantID, &out.OwnerID, &out.Bucket, &out.ObjectKey,
		&out.ContentType, &out.SizeBytes, &out.CreatedAt,
	)
	if err != nil {
		return File{}, fmt.Errorf("insert file: %w", err)
	}
	return out, nil
}

// GetByID loads file metadata by primary key.
func (s *Files) GetByID(ctx context.Context, id string) (File, error) {
	if s.pool == nil {
		return File{}, fmt.Errorf("database not connected")
	}
	const q = `
SELECT id, tenant_id, owner_id, bucket, object_key, content_type, size_bytes, created_at
FROM files
WHERE id = $1`

	var out File
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&out.ID, &out.TenantID, &out.OwnerID, &out.Bucket, &out.ObjectKey,
		&out.ContentType, &out.SizeBytes, &out.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return File{}, pgx.ErrNoRows
		}
		return File{}, fmt.Errorf("get file: %w", err)
	}
	return out, nil
}

// ListByOwner returns files owned by the user, newest first.
func (s *Files) ListByOwner(ctx context.Context, ownerID, tenantID string, limit int) ([]File, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
SELECT id, tenant_id, owner_id, bucket, object_key, content_type, size_bytes, created_at
FROM files
WHERE owner_id = $1
  AND ($2 = '' OR tenant_id = '' OR tenant_id = $2)
ORDER BY created_at DESC
LIMIT $3`

	rows, err := s.pool.Query(ctx, q, ownerID, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	var out []File
	for rows.Next() {
		var f File
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.OwnerID, &f.Bucket, &f.ObjectKey,
			&f.ContentType, &f.SizeBytes, &f.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	return out, nil
}

// DeleteByID removes metadata for an owned file. Returns pgx.ErrNoRows when missing or not owned.
func (s *Files) DeleteByID(ctx context.Context, id, ownerID, tenantID string) error {
	if s.pool == nil {
		return fmt.Errorf("database not connected")
	}
	const q = `
DELETE FROM files
WHERE id = $1
  AND owner_id = $2
  AND ($3 = '' OR tenant_id = '' OR tenant_id = $3)`

	tag, err := s.pool.Exec(ctx, q, id, ownerID, tenantID)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
