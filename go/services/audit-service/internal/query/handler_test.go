package query_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/audit-service/internal/query"
	"github.com/ting-boundless/boundless/services/audit-service/internal/store"
)

type stubEvents struct{}

func (stubEvents) List(context.Context, store.ListFilter) ([]store.EventRow, error) {
	return []store.EventRow{{
		ID: "e1", Source: "business-service", Type: "business.item.created",
		Time: time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC),
	}}, nil
}

func TestHandler_ListRequiresAdmin(t *testing.T) {
	h := query.New(stubEvents{})
	req := httptest.NewRequest(http.MethodGet, "/v1/audit/events", nil)
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", RequestID: "r1", Roles: []string{"user"},
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestHandler_ListOK(t *testing.T) {
	h := query.New(stubEvents{})
	req := httptest.NewRequest(http.MethodGet, "/v1/audit/events", nil)
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", TenantID: "t1", RequestID: "r1", Roles: []string{"admin"},
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "business.item.created") {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
