package main_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	pkgauth "github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/revocation"
	gwauth "github.com/ting-boundless/boundless/services/gateway/internal/auth"
	"github.com/ting-boundless/boundless/services/gateway/internal/proxy"
	"github.com/ting-boundless/boundless/services/gateway/internal/ratelimit"
)

func TestIntegration_RateLimit_AuthPath_429(t *testing.T) {
	called := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, err := proxy.New(proxy.Routes{"/v1/auth/": upstream.URL}, log, "")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", router)

	cfg := ratelimit.Config{
		AuthRPS:      1,
		AuthBurst:    1,
		GeneralRPS:   100,
		GeneralBurst: 100,
		Enabled:      true,
	}
	lim := ratelimit.NewLimiter(cfg, nil, log)
	v, err := pkgauth.NewVerifier(t.Context(), pkgauth.Config{DevSecret: "s"}, log)
	if err != nil {
		t.Fatal(err)
	}

	h := httpx.Chain(mux,
		ratelimit.Middleware(cfg, nil, lim),
		gwauth.Authenticate(v, nil, gwauth.DefaultAnonPrefixes(), nil, nil, nil, gwauth.SensitivePrefixes{}),
	)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/miniprogram/login", nil)
	req.RemoteAddr = "203.0.113.9:1234"

	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first status=%d body=%s", rr1.Code, rr1.Body.String())
	}

	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second status=%d body=%s", rr2.Code, rr2.Body.String())
	}
	if called != 1 {
		t.Fatalf("upstream called %d times, want 1", called)
	}
}

func TestIntegration_RevokedSubject_SensitiveFilesPath_401(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	revocations := revocation.NewStore(rdb)

	upstream := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("upstream should not be called for revoked subject on sensitive path")
	}))
	defer upstream.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := pkgauth.Config{
		Issuers:   []string{"http://test/oidc"},
		Audience:  "ting-test",
		DevSecret: "integration-secret",
	}
	verifier, err := pkgauth.NewVerifier(t.Context(), cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := pkgauth.DevToken(cfg, "revoked-user", "t1", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := revocations.RevokeSubject(t.Context(), "revoked-user"); err != nil {
		t.Fatal(err)
	}

	router, err := proxy.New(proxy.Routes{"/v1/files/": upstream.URL}, log, "")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", router)

	h := httpx.Chain(mux,
		gwauth.Authenticate(verifier, nil, gwauth.DefaultAnonPrefixes(), nil, nil, revocations, gwauth.DefaultSensitivePrefixes()),
	)

	req := httptest.NewRequest(http.MethodPost, "/v1/files/upload", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestIntegration_ClientRequestID_NotTrusted(t *testing.T) {
	const wantUser = "rid-user"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(identity.HeaderRequestID) == "client-forged-rid" {
			t.Error("gateway forwarded client request id")
		}
		if r.Header.Get(identity.HeaderRequestID) == "" {
			t.Error("missing regenerated request id")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := pkgauth.Config{DevSecret: "s"}
	verifier, err := pkgauth.NewVerifier(t.Context(), cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := pkgauth.DevToken(cfg, wantUser, "t1", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router, err := proxy.New(proxy.Routes{"/v1/users/": upstream.URL}, log, "")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", router)

	h := httpx.Chain(mux,
		gwauth.Authenticate(verifier, nil, gwauth.DefaultAnonPrefixes(), nil, nil, nil, gwauth.SensitivePrefixes{}),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set(identity.HeaderRequestID, "client-forged-rid")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if rr.Header().Get(identity.HeaderRequestID) == "client-forged-rid" {
		t.Fatal("response echoed client request id")
	}
	if rr.Header().Get(identity.HeaderRequestID) == "" {
		t.Fatal("response missing request id")
	}
}
