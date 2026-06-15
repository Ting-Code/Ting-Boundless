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

func TestIntegration_BearerJWT_Admin_UsersList(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get(identity.HeaderUserID) != "admin-user" {
			t.Errorf("X-User-Id=%q", r.Header.Get(identity.HeaderUserID))
		}
		if r.Header.Get(identity.HeaderTenantID) != "t1" {
			t.Errorf("X-Tenant-Id=%q", r.Header.Get(identity.HeaderTenantID))
		}
		httpx.JSON(w, http.StatusOK, map[string]any{
			"users": []map[string]any{{"user_id": "u1", "display_name": "Alice"}},
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
	tok, err := pkgauth.DevToken(cfg, "admin-user", "t1", []string{"admin"}, time.Hour)
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
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	var body struct {
		Users []map[string]any `json:"users"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Users) != 1 {
		t.Fatalf("users=%+v", body.Users)
	}
}
