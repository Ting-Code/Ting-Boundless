package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGatewayTrust_SkipsWhenTokenUnset(t *testing.T) {
	called := false
	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	rr := httptest.NewRecorder()
	GatewayTrust("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if !called || rr.Code != http.StatusOK {
		t.Fatalf("called=%v status=%d", called, rr.Code)
	}
}

func TestGatewayTrust_RejectsWithoutToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	rr := httptest.NewRecorder()
	GatewayTrust("secret")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not run")
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestGatewayTrust_AllowsProbePaths(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	GatewayTrust("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestGatewayTrust_AllowsWithInternalHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.Header.Set("X-Internal-Token", "secret")
	rr := httptest.NewRecorder()
	GatewayTrust("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}
