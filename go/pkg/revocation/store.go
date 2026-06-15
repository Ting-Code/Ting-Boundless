// Package revocation stores subject and session blocklists in Redis for Gateway checks.
package revocation

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

const defaultTTL = 30 * 24 * time.Hour

// Store tracks revoked JWT subjects and BFF session ids in Redis.
type Store struct {
	rdb    *goredis.Client
	prefix string
	ttl    time.Duration
}

// NewStore creates a revocation store. rdb may be nil (checks are skipped).
func NewStore(rdb *goredis.Client) *Store {
	ttl := defaultTTL
	if s := httpx.Env("AUTH_REVOCATION_TTL", ""); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			ttl = d
		}
	}
	return &Store{
		rdb:    rdb,
		prefix: httpx.Env("AUTH_REVOCATION_REDIS_PREFIX", "auth:revoked:"),
		ttl:    ttl,
	}
}

// Enabled reports whether Redis-backed revocation is available.
func (s *Store) Enabled() bool {
	return s != nil && s.rdb != nil
}

// RevokeSubject blocks a JWT subject (sub claim) until TTL expires.
func (s *Store) RevokeSubject(ctx context.Context, subject string) error {
	return s.set(ctx, "sub:"+subject)
}

// RevokeSession blocks a BFF session id until TTL expires.
func (s *Store) RevokeSession(ctx context.Context, sessionID string) error {
	return s.set(ctx, "session:"+sessionID)
}

// IsSubjectRevoked reports whether subject is on the blocklist.
func (s *Store) IsSubjectRevoked(ctx context.Context, subject string) (bool, error) {
	return s.exists(ctx, "sub:"+subject)
}

// IsSessionRevoked reports whether sessionID is on the blocklist.
func (s *Store) IsSessionRevoked(ctx context.Context, sessionID string) (bool, error) {
	return s.exists(ctx, "session:"+sessionID)
}

func (s *Store) set(ctx context.Context, suffix string) error {
	if !s.Enabled() || suffix == "" {
		return nil
	}
	key := s.prefix + suffix
	if s.ttl > 0 {
		return s.rdb.Set(ctx, key, "1", s.ttl).Err()
	}
	return s.rdb.Set(ctx, key, "1", 0).Err()
}

func (s *Store) exists(ctx context.Context, suffix string) (bool, error) {
	if !s.Enabled() || suffix == "" {
		return false, nil
	}
	n, err := s.rdb.Exists(ctx, s.prefix+suffix).Result()
	if err != nil {
		return false, fmt.Errorf("revocation lookup: %w", err)
	}
	return n > 0, nil
}
