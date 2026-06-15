package me_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/user-service/internal/me"
	"github.com/ting-boundless/boundless/services/user-service/internal/store"
)

type stubUsers struct {
	user store.User
}

func (s *stubUsers) GetOrCreate(context.Context, identity.Identity) (store.User, error) {
	return s.user, nil
}

func (s *stubUsers) UpdateDisplayName(_ context.Context, id identity.Identity, name string) (store.User, error) {
	u := s.user
	u.DisplayName = name
	u.TenantID = id.TenantID
	return u, nil
}

func TestHandler_GetMe(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC)
	h := me.New(&stubUsers{
		user: store.User{ID: "u1", TenantID: "t1", DisplayName: "Alice", CreatedAt: created, UpdatedAt: created},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", TenantID: "t1", Roles: []string{"admin"}, RequestID: "r1",
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "Alice") {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandler_PatchDisplayName(t *testing.T) {
	h := me.New(&stubUsers{
		user: store.User{ID: "u1", DisplayName: "old"},
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/users/me", strings.NewReader(`{"display_name":"新名字"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", RequestID: "r1",
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "新名字") {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandler_PatchRequiresDisplayName(t *testing.T) {
	h := me.New(&stubUsers{})

	req := httptest.NewRequest(http.MethodPatch, "/v1/users/me", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", RequestID: "r1",
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestHandler_Unauthenticated(t *testing.T) {
	h := me.New(&stubUsers{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/users/me", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}
