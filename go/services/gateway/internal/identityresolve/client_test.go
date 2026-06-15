package identityresolve_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/gateway/internal/identityresolve"
)

func TestClient_ResolveLogto(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/identity/resolve" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Internal-Token") != "secret" {
			t.Errorf("token=%q", r.Header.Get("X-Internal-Token"))
		}
		if r.Header.Get(identity.HeaderRequestID) != "rid-1" {
			t.Errorf("request id=%q", r.Header.Get(identity.HeaderRequestID))
		}
		var req struct {
			Provider    string `json:"provider"`
			ProviderUID string `json:"provider_uid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Provider != "logto" || req.ProviderUID != "sub-1" {
			t.Fatalf("req=%+v", req)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"user_id": "u-platform"})
	}))
	defer srv.Close()

	c := identityresolve.NewClient(srv.URL, "secret")
	userID, err := c.ResolveLogto(context.Background(), "sub-1", "rid-1")
	if err != nil {
		t.Fatal(err)
	}
	if userID != "u-platform" {
		t.Fatalf("user_id=%q", userID)
	}
}

func TestClient_NilReturnsError(t *testing.T) {
	var c *identityresolve.Client
	_, err := c.ResolveLogto(context.Background(), "sub", "rid")
	if err == nil {
		t.Fatal("expected error")
	}
}
