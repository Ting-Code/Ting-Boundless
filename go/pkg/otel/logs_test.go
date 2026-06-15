package otel

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestAttachLogExport_DisabledWithoutEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	base := slog.New(slog.NewTextHandler(os.Stderr, nil))
	out, shutdown, err := AttachLogExport(context.Background(), "test", base)
	if err != nil {
		t.Fatal(err)
	}
	if out != base {
		t.Fatal("expected same logger when disabled")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestAttachLogExport_DisabledWhenExporterNone(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4317")
	t.Setenv("OTEL_LOGS_EXPORTER", "none")
	base := slog.New(slog.NewTextHandler(os.Stderr, nil))
	out, _, err := AttachLogExport(context.Background(), "test", base)
	if err != nil {
		t.Fatal(err)
	}
	if out != base {
		t.Fatal("expected same logger when OTEL_LOGS_EXPORTER=none")
	}
}

func TestLogsExportEnabled(t *testing.T) {
	t.Setenv("OTEL_LOGS_EXPORTER", "")
	if !logsExportEnabled() {
		t.Fatal("default should enable logs")
	}
	t.Setenv("OTEL_LOGS_EXPORTER", "none")
	if logsExportEnabled() {
		t.Fatal("none should disable")
	}
}
