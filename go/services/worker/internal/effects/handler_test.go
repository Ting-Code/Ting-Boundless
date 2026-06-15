package effects_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ting-boundless/boundless/services/worker/internal/effects"
	"github.com/ting-boundless/boundless/services/worker/internal/store"
)

type stubLister struct{}

func (stubLister) List(context.Context, store.ListFilter) ([]store.EffectRow, error) {
	return []store.EffectRow{{
		JobID: "j1", JobType: "business.item.created", ResourceID: "item-1",
		ProcessedAt: time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC),
	}}, nil
}

func TestHandler_List(t *testing.T) {
	h := effects.New(stubLister{})
	req := httptest.NewRequest(http.MethodGet, "/internal/job-effects?limit=10", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "business.item.created") {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
