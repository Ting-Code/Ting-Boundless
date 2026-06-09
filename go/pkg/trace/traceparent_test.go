package trace

import "testing"

func TestTraceIDFromParent(t *testing.T) {
	tp := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	if got := TraceIDFromParent(tp); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace id: got %q", got)
	}
	if TraceIDFromParent("") != "" {
		t.Fatal("expected empty for missing")
	}
	if TraceIDFromParent("invalid") != "" {
		t.Fatal("expected empty for invalid")
	}
}

func TestNewTraceparent(t *testing.T) {
	tp := NewTraceparent()
	if TraceIDFromParent(tp) == "" {
		t.Fatalf("invalid traceparent: %q", tp)
	}
}
