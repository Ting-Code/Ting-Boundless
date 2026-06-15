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

func TestIntegration_BearerJWT_Admin_AuditEvents(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audit/events" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get(identity.HeaderUserID) == "" {
			t.Error("upstream missing X-User-Id")
		}
		if r.Header.Get("X-Internal-Token") != "test-internal" {
			t.Errorf("upstream internal token=%q", r.Header.Get("X-Internal-Token"))
		}
		httpx.JSON(w, http.StatusOK, map[string]any{
			"events": []map[string]any{{
				"id": "e1", "source": "business-service", "type": "business.item.created",
			}},
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

	router, err := proxy.New(proxy.Routes{"/v1/audit/": upstream.URL}, log, "test-internal")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)

	h := httpx.Chain(mux,
		gwauth.Authenticate(verifier, nil, gwauth.DefaultAnonPrefixes(), nil, nil, nil, gwauth.SensitivePrefixes{}),
		httpx.Recover(log),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/events?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	var body struct {
		Events []map[string]any `json:"events"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Events) != 1 {
		t.Fatalf("events=%+v", body.Events)
	}
}
