package otel

import (
	"context"
	"testing"
)

func TestInitFromEnv_DisabledWhenUnset(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	shutdown, err := InitFromEnv(context.Background(), "test-svc", nil)
	if err != nil {
		t.Fatal(err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown func")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestInitFromEnv_DisabledForPlaceholder(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-placeholder:4317")
	shutdown, err := InitFromEnv(context.Background(), "test-svc", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}
