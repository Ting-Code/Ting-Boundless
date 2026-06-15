package httpx

import (
	"io"
	"log/slog"
	"testing"
)

func discardLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRequireInternalToken(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("REQUIRE_INTERNAL_TOKEN", "")
	if RequireInternalToken() {
		t.Fatal("dev default should not require")
	}

	t.Setenv("APP_ENV", "production")
	if !RequireInternalToken() {
		t.Fatal("production should require")
	}

	t.Setenv("APP_ENV", "development")
	t.Setenv("REQUIRE_INTERNAL_TOKEN", "true")
	if !RequireInternalToken() {
		t.Fatal("explicit flag should require")
	}
}

func TestLoadInternalToken_DevOptional(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("REQUIRE_INTERNAL_TOKEN", "")
	t.Setenv("INTERNAL_API_TOKEN", "")

	token, ok := LoadInternalToken(discardLog())
	if !ok || token != "" {
		t.Fatalf("ok=%v token=%q", ok, token)
	}
}

func TestLoadInternalToken_ProductionRequired(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("INTERNAL_API_TOKEN", "")

	_, ok := LoadInternalToken(discardLog())
	if ok {
		t.Fatal("expected failure when production and token unset")
	}
}

func TestLoadInternalToken_ReturnsValue(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("INTERNAL_API_TOKEN", "  secret  ")

	token, ok := LoadInternalToken(discardLog())
	if !ok || token != "secret" {
		t.Fatalf("ok=%v token=%q", ok, token)
	}
}
