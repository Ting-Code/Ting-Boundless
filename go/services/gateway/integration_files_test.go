package main_test

import (
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

func TestIntegration_BearerJWT_FilesDelete(t *testing.T) {
	const wantUser = "file-owner"
	const fileID = "f-delete-me"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/files/"+fileID {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get(identity.HeaderUserID); got != wantUser {
			t.Errorf("upstream X-User-Id=%q want %q", got, wantUser)
		}
		if r.Header.Get("X-Internal-Token") != "test-internal" {
			t.Errorf("upstream internal token=%q", r.Header.Get("X-Internal-Token"))
		}
		w.WriteHeader(http.StatusNoContent)
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

	router, err := proxy.New(proxy.Routes{"/v1/files/": upstream.URL}, log, "test-internal")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)

	h := httpx.Chain(mux,
		gwauth.Authenticate(verifier, nil, gwauth.DefaultAnonPrefixes(), nil, nil, nil, gwauth.SensitivePrefixes{}),
		httpx.Recover(log),
	)

	req := httptest.NewRequest(http.MethodDelete, "/v1/files/"+fileID, nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
