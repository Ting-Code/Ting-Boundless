// Package ratelimit provides per-client IP rate limiting at the Gateway edge.
package ratelimit

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/time/rate"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// Config holds token-bucket limits (requests per second, burst).
type Config struct {
	AuthRPS      float64
	AuthBurst    int
	GeneralRPS   float64
	GeneralBurst int
	Enabled      bool
}

// ConfigFromEnv loads limits from the environment.
func ConfigFromEnv() Config {
	return Config{
		AuthRPS:      envFloat("GATEWAY_RATE_LIMIT_AUTH_RPS", 5),
		AuthBurst:    envInt("GATEWAY_RATE_LIMIT_AUTH_BURST", 10),
		GeneralRPS:   envFloat("GATEWAY_RATE_LIMIT_GENERAL_RPS", 50),
		GeneralBurst: envInt("GATEWAY_RATE_LIMIT_GENERAL_BURST", 100),
		Enabled:      httpx.Env("GATEWAY_RATE_LIMIT_ENABLED", "true") != "false",
	}
}

func isAuthPath(path string) bool {
	return strings.HasPrefix(path, "/v1/auth/") ||
		path == "/internal/webhooks/logto" ||
		strings.HasPrefix(path, "/sign-in") ||
		path == "/callback"
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type pool struct {
	mu    sync.Mutex
	lim   rate.Limit
	burst int
	byKey map[string]*rate.Limiter
}

func newPool(rps float64, burst int) *pool {
	if rps <= 0 {
		return nil
	}
	if burst <= 0 {
		burst = int(rps)
		if burst < 1 {
			burst = 1
		}
	}
	return &pool{
		lim:   rate.Limit(rps),
		burst: burst,
		byKey: make(map[string]*rate.Limiter),
	}
}

func (p *pool) allow(key string) bool {
	if p == nil {
		return true
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	lim, ok := p.byKey[key]
	if !ok {
		lim = rate.NewLimiter(p.lim, p.burst)
		p.byKey[key] = lim
	}
	return lim.Allow()
}

func envFloat(key string, def float64) float64 {
	v := httpx.Env(key, "")
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f < 0 {
		return def
	}
	return f
}

func envInt(key string, def int) int {
	v := httpx.Env(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}
