package ingest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/services/audit-service/internal/ingest"
)

type stubWriter struct {
	last audit.Event
}

func (s *stubWriter) Insert(_ context.Context, e audit.Event) error {
	s.last = e
	return nil
}

func TestHandler_AcceptsValidEvent(t *testing.T) {
	w := &stubWriter{}
	h := ingest.New(w)
	body := `{"id":"e1","source":"gateway","type":"api.access.denied","time":"2026-06-05T12:00:00Z","data":{"path":"/v1/x"}}`
	req := httptest.NewRequest(http.MethodPost, "/internal/audit/events", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if w.last.ID != "e1" {
		t.Fatalf("got id %q", w.last.ID)
	}
}

func TestHandler_RejectsMissingFields(t *testing.T) {
	h := ingest.New(&stubWriter{})
	req := httptest.NewRequest(http.MethodPost, "/internal/audit/events", strings.NewReader(`{"id":"e3"}`))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

// Ensure time import used in valid test indirectly via JSON.
var _ = time.UTC
