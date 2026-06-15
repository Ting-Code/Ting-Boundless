package ratelimit

import (
	"context"
	"log/slog"
	"net/http"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// Limiter checks per-client IP quotas.
type Limiter interface {
	Allow(ctx context.Context, zone, ip string) bool
}

type memoryLimiter struct {
	auth    *pool
	general *pool
}

func newMemoryLimiter(cfg Config) *memoryLimiter {
	return &memoryLimiter{
		auth:    newPool(cfg.AuthRPS, cfg.AuthBurst),
		general: newPool(cfg.GeneralRPS, cfg.GeneralBurst),
	}
}

func (m *memoryLimiter) Allow(_ context.Context, zone, ip string) bool {
	p := m.general
	if zone == "auth" {
		p = m.auth
	}
	if p == nil {
		return true
	}
	return p.allow(ip)
}

type redisLimiterFacade struct {
	inner *redisLimiter
	log   *slog.Logger
	mem   *memoryLimiter
}

func (r *redisLimiterFacade) Allow(ctx context.Context, zone, ip string) bool {
	ok, err := r.inner.allow(ctx, zone, ip)
	if err != nil {
		if r.log != nil {
			r.log.Warn("redis rate limit failed; falling back to memory", slog.Any("error", err))
		}
		return r.mem.Allow(ctx, zone, ip)
	}
	return ok
}

// NewLimiter returns a Redis-backed limiter when available, else in-memory.
func NewLimiter(cfg Config, rdb *goredis.Client, log *slog.Logger) Limiter {
	mem := newMemoryLimiter(cfg)
	if rdb == nil || httpx.Env("GATEWAY_RATE_LIMIT_REDIS", "true") == "false" {
		return mem
	}
	return &redisLimiterFacade{
		inner: newRedisLimiter(rdb, cfg),
		log:   log,
		mem:   mem,
	}
}

// Middleware applies per-IP limits. Auth paths use the stricter bucket.
func Middleware(cfg Config, entryAudit audit.Emitter, limiter Limiter) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	if limiter == nil {
		limiter = newMemoryLimiter(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			zone := "general"
			if isAuthPath(r.URL.Path) {
				zone = "auth"
			}
			if !limiter.Allow(r.Context(), zone, ip) {
				rid := httpx.NewRequestID()
				w.Header().Set("Retry-After", "1")
				audit.EmitEntry(entryAudit, r, rid, "api.rate_limited", map[string]any{
					"zone":      zone,
					"client_ip": ip,
				})
				httpx.WriteError(w, rid, errs.New(http.StatusTooManyRequests, "rate_limited", "too many requests"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
