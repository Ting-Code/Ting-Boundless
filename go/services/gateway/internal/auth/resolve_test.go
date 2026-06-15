package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/gateway/internal/identityresolve"
)

func TestMapPlatformUserID_SkipsWhenDistinctUserID(t *testing.T) {
	id := identity.Identity{UserID: "platform-1", Subject: "logto-sub"}
	got := mapPlatformUserID(context.Background(), identityresolve.NewClient("http://unused", ""), id, "rid")
	if got.UserID != "platform-1" {
		t.Fatalf("user_id=%q", got.UserID)
	}
}

func TestMapPlatformUserID_ResolvesLogtoSubject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/identity/resolve" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"user_id": "resolved-platform-id"})
	}))
	defer srv.Close()

	client := identityresolve.NewClient(srv.URL, "")
	id := identity.Identity{UserID: "logto-sub-abc", Subject: "logto-sub-abc"}
	got := mapPlatformUserID(context.Background(), client, id, "rid")
	if got.UserID != "resolved-platform-id" {
		t.Fatalf("user_id=%q want resolved-platform-id", got.UserID)
	}
}

func TestMapPlatformUserID_NilResolverKeepsSubject(t *testing.T) {
	id := identity.Identity{UserID: "sub-only", Subject: "sub-only"}
	got := mapPlatformUserID(context.Background(), nil, id, "rid")
	if got.UserID != "sub-only" {
		t.Fatalf("user_id=%q", got.UserID)
	}
}
