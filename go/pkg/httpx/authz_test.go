package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
)

func TestRequireRole_AllowsAdmin(t *testing.T) {
	var called bool
	h := httpx.TrustedRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	identity.Identity{
		UserID: "u1", Roles: []string{"admin"}, RequestID: "r1",
	}.Inject(req.Header)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called || rr.Code != http.StatusNoContent {
		t.Fatalf("called=%v status=%d", called, rr.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	h := httpx.TrustedRole("admin")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("should not run")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	identity.Identity{
		UserID: "u1", Roles: []string{"user"}, RequestID: "r1",
	}.Inject(req.Header)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
}
