package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ting-boundless/boundless/pkg/trace"
)

func TestTraceContext_PreservesIncomingTraceparent(t *testing.T) {
	in := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(trace.HeaderTraceparent, in)

	rr := httptest.NewRecorder()
	TraceContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(trace.HeaderTraceparent); got != in {
			t.Fatalf("request traceparent=%q want %q", got, in)
		}
	})).ServeHTTP(rr, req)

	if got := rr.Header().Get(trace.HeaderTraceparent); got != in {
		t.Fatalf("response traceparent=%q want %q", got, in)
	}
}

func TestTraceContext_GeneratesWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	TraceContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(trace.HeaderTraceparent) == "" {
			t.Fatal("expected generated traceparent on request")
		}
	})).ServeHTTP(rr, req)

	if trace.TraceIDFromParent(rr.Header().Get(trace.HeaderTraceparent)) == "" {
		t.Fatal("expected valid traceparent on response")
	}
}
