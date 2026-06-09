package auth

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/gateway/internal/session"
)

func TestAuthenticate_StripsClientHeadersAndVerifiesJWT(t *testing.T) {
	cfg := auth.Config{
		Issuer:    "http://test/oidc",
		Audience:  "ting-test",
		DevSecret: "edge-secret",
	}
	v, err := auth.NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}

	tok, err := auth.DevToken(cfg, "real-user", "t1", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	var got identity.Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = identity.FromHeaders(r.Header)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.Header.Set("X-User-Id", "forged")
	req.Header.Set("X-Request-Id", "client-rid")
	req.Header.Set("Authorization", "Bearer "+tok)

	rr := httptest.NewRecorder()
	Authenticate(v, nil, DefaultAnonPrefixes())(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	if got.UserID != "real-user" {
		t.Fatalf("user_id=%q", got.UserID)
	}
	if got.RequestID == "" || got.RequestID == "client-rid" {
		t.Fatalf("request_id should be regenerated, got %q", got.RequestID)
	}
}

func TestAuthenticate_InvalidBearerReturns401(t *testing.T) {
	cfg := auth.Config{DevSecret: "s"}
	v, err := auth.NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer bad")

	rr := httptest.NewRecorder()
	Authenticate(v, nil, DefaultAnonPrefixes())(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestAuthenticate_SessionCookie(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	sessions := session.NewStore(rdb)

	cfg := auth.Config{
		Issuer:    "http://test/oidc",
		Audience:  "ting-test",
		DevSecret: "cookie-secret",
	}
	v, err := auth.NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}

	tok, err := auth.DevToken(cfg, "cookie-user", "t1", []string{"admin"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id, err := v.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}

	sid, err := sessions.Create(t.Context(), session.Data{
		AccessToken: tok,
		Identity:    id,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	var got identity.Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = identity.FromHeaders(r.Header)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.AddCookie(&http.Cookie{Name: sessions.CookieName(), Value: sid})

	rr := httptest.NewRecorder()
	Authenticate(v, sessions, DefaultAnonPrefixes())(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	if got.UserID != "cookie-user" {
		t.Fatalf("user_id=%q", got.UserID)
	}
}

func TestAuthenticate_NonAnonWithoutCredentialsReturns401(t *testing.T) {
	cfg := auth.Config{DevSecret: "s"}
	v, err := auth.NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/business/me", nil)
	rr := httptest.NewRecorder()
	Authenticate(v, nil, DefaultAnonPrefixes())(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not run")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}

func TestAuthenticate_AnonPathAllowsWithoutCredentials(t *testing.T) {
	cfg := auth.Config{DevSecret: "s"}
	v, err := auth.NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}

	called := false
	req := httptest.NewRequest(http.MethodGet, "/v1/business/ping", nil)
	rr := httptest.NewRecorder()
	Authenticate(v, nil, DefaultAnonPrefixes())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	if !called {
		t.Fatal("expected next handler to run for anonymous whitelist path")
	}
}
