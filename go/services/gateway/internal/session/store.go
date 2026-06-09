// Package session stores Gateway BFF sessions in Redis (HttpOnly cookie → session id).
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
)

// Data is the server-side session payload (tokens never exposed to browser JS).
type Data struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	Identity     identity.Identity `json:"identity"`
	ExpiresAt    time.Time         `json:"expires_at"`
}

// Store persists BFF sessions and OIDC state in Redis.
type Store struct {
	rdb        *goredis.Client
	ttl        time.Duration
	cookieName string
	prefix     string
}

// NewStore creates a session store. rdb may be nil (cookie auth disabled).
func NewStore(rdb *goredis.Client) *Store {
	ttl := 24 * time.Hour
	if s := httpx.Env("SESSION_TTL", ""); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			ttl = d
		}
	}
	return &Store{
		rdb:        rdb,
		ttl:        ttl,
		cookieName: httpx.Env("SESSION_COOKIE_NAME", "tb_session"),
		prefix:     httpx.Env("SESSION_REDIS_PREFIX", "gw:session:"),
	}
}

// Enabled reports whether Redis-backed sessions are available.
func (s *Store) Enabled() bool {
	return s != nil && s.rdb != nil
}

// CookieName returns the HttpOnly session cookie name.
func (s *Store) CookieName() string {
	if s == nil {
		return "tb_session"
	}
	return s.cookieName
}

// Create stores a new session and returns its id.
func (s *Store) Create(ctx context.Context, data Data) (string, error) {
	if !s.Enabled() {
		return "", fmt.Errorf("session store unavailable")
	}
	id, err := newID()
	if err != nil {
		return "", err
	}
	if err := s.save(ctx, id, data); err != nil {
		return "", err
	}
	return id, nil
}

// Get loads a session by id.
func (s *Store) Get(ctx context.Context, id string) (Data, error) {
	if !s.Enabled() {
		return Data{}, fmt.Errorf("session store unavailable")
	}
	if id == "" {
		return Data{}, fmt.Errorf("empty session id")
	}
	raw, err := s.rdb.Get(ctx, s.prefix+id).Bytes()
	if err != nil {
		return Data{}, err
	}
	var data Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return Data{}, err
	}
	if !data.ExpiresAt.IsZero() && time.Now().After(data.ExpiresAt) {
		_ = s.Delete(ctx, id)
		return Data{}, fmt.Errorf("session expired")
	}
	return data, nil
}

// Delete removes a session.
func (s *Store) Delete(ctx context.Context, id string) error {
	if !s.Enabled() || id == "" {
		return nil
	}
	return s.rdb.Del(ctx, s.prefix+id).Err()
}

func (s *Store) save(ctx context.Context, id string, data Data) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ttl := s.ttl
	if !data.ExpiresAt.IsZero() {
		if d := time.Until(data.ExpiresAt); d > 0 && d < ttl {
			ttl = d
		}
	}
	return s.rdb.Set(ctx, s.prefix+id, raw, ttl).Err()
}

// PendingLogin holds OIDC state between /sign-in and /callback.
type PendingLogin struct {
	ReturnTo string `json:"return_to"`
	Nonce    string `json:"nonce"`
}

// SavePending stores OIDC state (short TTL).
func (s *Store) SavePending(ctx context.Context, state string, p PendingLogin) error {
	if !s.Enabled() {
		return fmt.Errorf("session store unavailable")
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, "gw:oidc:state:"+state, raw, 10*time.Minute).Err()
}

// ConsumePending loads and deletes OIDC state.
func (s *Store) ConsumePending(ctx context.Context, state string) (PendingLogin, error) {
	if !s.Enabled() {
		return PendingLogin{}, fmt.Errorf("session store unavailable")
	}
	key := "gw:oidc:state:" + state
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return PendingLogin{}, err
	}
	_ = s.rdb.Del(ctx, key).Err()
	var p PendingLogin
	if err := json.Unmarshal(raw, &p); err != nil {
		return PendingLogin{}, err
	}
	return p, nil
}

func newID() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
