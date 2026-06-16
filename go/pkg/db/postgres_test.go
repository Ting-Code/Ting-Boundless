package db_test

import (
	"os"
	"testing"

	"github.com/ting-boundless/boundless/pkg/db"
)

func TestAuditConfigFromEnv_OverridesUser(t *testing.T) {
	t.Setenv("POSTGRES_USER", "ting")
	t.Setenv("POSTGRES_PASSWORD", "ting-pass")
	t.Setenv("AUDIT_POSTGRES_USER", "audit_writer")
	t.Setenv("AUDIT_POSTGRES_PASSWORD", "audit-pass")

	cfg := db.AuditConfigFromEnv("audit_db")
	if cfg.User != "audit_writer" || cfg.Password != "audit-pass" || cfg.Database != "audit_db" {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestAuditConfigFromEnv_FallsBackToPostgres(t *testing.T) {
	os.Unsetenv("AUDIT_POSTGRES_USER")
	os.Unsetenv("AUDIT_POSTGRES_PASSWORD")
	t.Setenv("POSTGRES_USER", "ting")
	t.Setenv("POSTGRES_PASSWORD", "change-me")

	cfg := db.AuditConfigFromEnv("audit_db")
	if cfg.User != "ting" || cfg.Password != "change-me" {
		t.Fatalf("cfg=%+v", cfg)
	}
}
