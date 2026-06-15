package main_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgauth "github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	gwauth "github.com/ting-boundless/boundless/services/gateway/internal/auth"
	"github.com/ting-boundless/boundless/services/gateway/internal/proxy"
)

// Q-02: gateway verifies Bearer JWT, injects trusted headers, proxies to user-service.
func TestIntegration_BearerJWT_UsersMe(t *testing.T) {
	const wantUser = "integration-user"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get(identity.HeaderUserID); got != wantUser {
			t.Errorf("upstream X-User-Id=%q want %q", got, wantUser)
		}
		if r.Header.Get(identity.HeaderRequestID) == "" {
			t.Error("upstream missing X-Request-Id")
		}
		if r.Header.Get("X-Internal-Token") != "test-internal" {
			t.Errorf("upstream internal token=%q", r.Header.Get("X-Internal-Token"))
		}
		httpx.JSON(w, http.StatusOK, map[string]any{
			"user_id":   wantUser,
			"tenant_id": "t1",
		})
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
	tok, err := pkgauth.DevToken(cfg, wantUser, "t1", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router, err := proxy.New(proxy.Routes{"/v1/users/": upstream.URL}, log, "test-internal")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)

	h := httpx.Chain(mux,
		gwauth.Authenticate(verifier, nil, gwauth.DefaultAnonPrefixes(), nil, nil, nil, gwauth.SensitivePrefixes{}),
		httpx.Recover(log),
		httpx.TraceContext,
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set(identity.HeaderUserID, "forged")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rid := rr.Header().Get(identity.HeaderRequestID); rid == "" {
		t.Fatal("response missing X-Request-Id")
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["user_id"] != wantUser {
		t.Fatalf("body user_id=%v", body["user_id"])
	}
}

func TestIntegration_Unauthenticated_UsersMe_401(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("upstream should not be called")
	}))
	defer upstream.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := pkgauth.Config{DevSecret: "s"}
	verifier, err := pkgauth.NewVerifier(t.Context(), cfg, log)
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
		httpx.Recover(log),
	)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/users/me", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
