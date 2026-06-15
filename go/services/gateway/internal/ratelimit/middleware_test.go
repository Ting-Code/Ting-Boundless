package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/ting-boundless/boundless/pkg/audit"
)

type syncAudit struct {
	mu     sync.Mutex
	events []audit.Event
}

func (s *syncAudit) Emit(_ context.Context, e audit.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
	return nil
}

func TestMiddleware_AuthPathRateLimited(t *testing.T) {
	aud := &syncAudit{}
	cfg := Config{AuthRPS: 1, AuthBurst: 1, GeneralRPS: 100, GeneralBurst: 100, Enabled: true}
	lim := newMemoryLimiter(cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := Middleware(cfg, aud, lim)(next)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/miniprogram/login", nil)
	req.RemoteAddr = "203.0.113.1:1234"

	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first status %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second status %d", rr2.Code)
	}
	aud.mu.Lock()
	defer aud.mu.Unlock()
	if len(aud.events) != 1 || aud.events[0].Type != "api.rate_limited" {
		t.Fatalf("events=%+v", aud.events)
	}
}

func TestMiddleware_DisabledPassthrough(t *testing.T) {
	cfg := Config{Enabled: false}
	called := false
	h := Middleware(cfg, nil, nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/v1/auth/x", nil))
	if !called {
		t.Fatal("expected next handler")
	}
}

func TestRedisLimiter_SharedAcrossInstances(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	cfg := Config{AuthRPS: 5, AuthBurst: 2, GeneralRPS: 100, GeneralBurst: 100, Enabled: true}
	lim := newRedisLimiter(rdb, cfg)
	ctx := context.Background()

	if !mustAllow(t, lim, ctx, "auth", "1.2.3.4") {
		t.Fatal("first")
	}
	if !mustAllow(t, lim, ctx, "auth", "1.2.3.4") {
		t.Fatal("second")
	}
	if mustAllow(t, lim, ctx, "auth", "1.2.3.4") {
		t.Fatal("third should be denied with burst=2")
	}
}

func mustAllow(t *testing.T, lim *redisLimiter, ctx context.Context, zone, ip string) bool {
	t.Helper()
	ok, err := lim.allow(ctx, zone, ip)
	if err != nil {
		t.Fatal(err)
	}
	return ok
}
