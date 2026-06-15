package list_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/user-service/internal/list"
	"github.com/ting-boundless/boundless/services/user-service/internal/store"
)

type stubLister struct{}

func (stubLister) ListByTenant(context.Context, string, int) ([]store.User, error) {
	return []store.User{{
		ID: "u1", TenantID: "t1", DisplayName: "Alice",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}, nil
}

func adminHandler() http.Handler {
	return identity.Middleware(httpx.RequireRole("admin")(list.New(stubLister{})))
}

func TestHandler_ListRequiresAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/users/", nil)
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", RequestID: "r1",
	}))
	rr := httptest.NewRecorder()
	adminHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestHandler_ListOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/users/", nil)
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", TenantID: "t1", Roles: []string{"admin"}, RequestID: "r1",
	}))
	rr := httptest.NewRecorder()
	adminHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "Alice") {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
