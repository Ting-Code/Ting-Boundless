package auth

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestVerifier_DevHMAC(t *testing.T) {
	cfg := Config{
		Issuers:   []string{"http://test/oidc"},
		Audience:  "ting-test",
		DevSecret: "test-secret",
	}
	v, err := NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}
	if !v.Enabled() {
		t.Fatal("expected verifier enabled")
	}

	tok, err := DevToken(cfg, "user-42", "tenant-1", []string{"admin"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id, err := v.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if id.UserID != "user-42" || id.TenantID != "tenant-1" || id.Subject != "user-42" {
		t.Fatalf("unexpected identity: %+v", id)
	}
	if len(id.Roles) != 1 || id.Roles[0] != "admin" {
		t.Fatalf("roles: %v", id.Roles)
	}
}

func TestVerifier_InvalidToken(t *testing.T) {
	cfg := Config{DevSecret: "secret"}
	v, err := NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Verify("not-a-jwt"); err == nil {
		t.Fatal("expected error for garbage token")
	}
}
