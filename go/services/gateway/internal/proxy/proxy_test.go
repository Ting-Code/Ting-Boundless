package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ting-boundless/boundless/services/gateway/internal/proxy"
)

func TestRouter_NotFoundUsesErrorEnvelope(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer up.Close()

	r, err := proxy.New(proxy.Routes{"/v1/users/": up.URL}, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	req.Header.Set("X-Request-Id", "rid-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "route.not_found") {
		t.Fatalf("body=%s", rr.Body.String())
	}
}
