package db

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// Postgres wraps a pgx connection pool.
type Postgres struct {
	pool *pgxpool.Pool
}

// Config holds PostgreSQL connection parameters (12-Factor via env).
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// ConfigFromEnv builds Config from standard POSTGRES_* env vars.
// database overrides POSTGRES_DB when non-empty (e.g. audit_db for audit-service).
func ConfigFromEnv(database string) Config {
	if database == "" {
		database = httpx.Env("POSTGRES_DB", "app_db")
	}
	return Config{
		Host:     httpx.Env("POSTGRES_HOST", "127.0.0.1"),
		Port:     httpx.Env("POSTGRES_PORT", "5432"),
		User:     httpx.Env("POSTGRES_USER", "ting"),
		Password: httpx.Env("POSTGRES_PASSWORD", ""),
		Database: database,
		SSLMode:  httpx.Env("POSTGRES_SSLMODE", "disable"),
	}
}

// DSN returns a PostgreSQL connection URI.
func (c Config) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   c.Database,
	}
	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

// Open connects to PostgreSQL and verifies connectivity with Ping.
// Returns an error when host is a cloud placeholder or the server is unreachable.
func Open(ctx context.Context, cfg Config) (*Postgres, error) {
	if config.IsPlaceholder(cfg.Host) {
		return nil, fmt.Errorf("postgres host is placeholder (%q)", cfg.Host)
	}
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

// Ping checks database connectivity.
func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Close shuts down the connection pool.
func (p *Postgres) Close() {
	p.pool.Close()
}

// Pool exposes the underlying pgx pool for queries.
func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}
