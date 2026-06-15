package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	providerWeChatMP    = "wechat_mp"
	providerWeChatUnion = "wechat_union"
	// ProviderLogto is the identity provider key for Logto OIDC subjects (JWT sub).
	ProviderLogto = "logto"
)

// IdentityStore resolves WeChat openids and unionids to platform user IDs.
type IdentityStore struct {
	pool *pgxpool.Pool
}

// NewIdentityStore wraps a Postgres pool.
func NewIdentityStore(pool *pgxpool.Pool) *IdentityStore {
	return &IdentityStore{pool: pool}
}

// ResolveProviderUser returns an existing or newly created platform user_id for
// an external identity provider subject (e.g. Logto sub).
func (s *IdentityStore) ResolveProviderUser(ctx context.Context, provider, providerUID string) (string, error) {
	if s == nil || s.pool == nil {
		return "", fmt.Errorf("identity store not configured")
	}
	if provider == "" || providerUID == "" {
		return "", fmt.Errorf("provider and provider_uid required")
	}

	if userID, err := s.lookupUserID(ctx, provider, providerUID); err == nil {
		return userID, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	userID, err := s.createUserWithProvider(ctx, provider, providerUID)
	if err != nil && isUniqueViolation(err) {
		if userID, err := s.lookupUserID(ctx, provider, providerUID); err == nil {
			return userID, nil
		}
	}
	return userID, err
}

// ResolveLogtoUser maps a Logto OIDC subject to a platform user_id.
func (s *IdentityStore) ResolveLogtoUser(ctx context.Context, logtoSub string) (string, error) {
	return s.ResolveProviderUser(ctx, ProviderLogto, logtoSub)
}

// ResolveWeChatUser returns an existing or newly created user_id.
// When unionid is present, identities from different mini-programs share one user_id.
func (s *IdentityStore) ResolveWeChatUser(ctx context.Context, openid, unionid string) (string, error) {
	if s == nil || s.pool == nil {
		return "", fmt.Errorf("identity store not configured")
	}
	if openid == "" {
		return "", fmt.Errorf("openid required")
	}

	if userID, err := s.lookupUserID(ctx, providerWeChatMP, openid); err == nil {
		if unionid != "" {
			if err := s.linkIdentity(ctx, providerWeChatUnion, unionid, userID); err != nil {
				return "", err
			}
		}
		return userID, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	if unionid != "" {
		if userID, err := s.lookupUserID(ctx, providerWeChatUnion, unionid); err == nil {
			if err := s.linkIdentity(ctx, providerWeChatMP, openid, userID); err != nil {
				return "", err
			}
			return userID, nil
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return "", err
		}
	}

	userID, err := s.createWeChatUser(ctx, openid, unionid)
	if err != nil && isUniqueViolation(err) {
		if userID, err := s.lookupUserID(ctx, providerWeChatMP, openid); err == nil {
			return userID, nil
		}
		if unionid != "" {
			if userID, err := s.lookupUserID(ctx, providerWeChatUnion, unionid); err == nil {
				_ = s.linkIdentity(ctx, providerWeChatMP, openid, userID)
				return userID, nil
			}
		}
	}
	return userID, err
}

func (s *IdentityStore) lookupUserID(ctx context.Context, provider, uid string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx,
		`SELECT user_id FROM user_identities WHERE provider = $1 AND provider_uid = $2`,
		provider, uid,
	).Scan(&userID)
	return userID, err
}

func (s *IdentityStore) linkIdentity(ctx context.Context, provider, uid, userID string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO user_identities (provider, provider_uid, user_id, tenant_id)
		 VALUES ($1, $2, $3, '')
		 ON CONFLICT (provider, provider_uid) DO NOTHING`,
		provider, uid, userID,
	)
	return err
}

func (s *IdentityStore) createUserWithProvider(ctx context.Context, provider, providerUID string) (string, error) {
	userID, err := newUserID()
	if err != nil {
		return "", err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`INSERT INTO users (id, tenant_id, display_name) VALUES ($1, '', '')`,
		userID,
	); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO user_identities (provider, provider_uid, user_id, tenant_id) VALUES ($1, $2, $3, '')`,
		provider, providerUID, userID,
	); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return userID, nil
}

func (s *IdentityStore) createWeChatUser(ctx context.Context, openid, unionid string) (string, error) {
	userID, err := s.createUserWithProvider(ctx, providerWeChatMP, openid)
	if err != nil {
		return "", err
	}
	if unionid != "" {
		if err := s.linkIdentity(ctx, providerWeChatUnion, unionid, userID); err != nil {
			return "", err
		}
	}
	return userID, nil
}

func newUserID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}
