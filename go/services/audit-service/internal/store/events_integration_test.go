//go:build integration

package store_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/services/audit-service/internal/store"
)

func TestAuditWriter_AppendOnly(t *testing.T) {
	auditDB := os.Getenv("AUDIT_DB")
	if auditDB == "" {
		auditDB = "audit_db"
	}

	runtimeCfg := db.AuditConfigFromEnv(auditDB)
	if runtimeCfg.Password == "" {
		t.Skip("AUDIT_POSTGRES_PASSWORD or POSTGRES_PASSWORD required")
	}

	ctx := context.Background()
	pg, err := db.Open(ctx, runtimeCfg)
	if err != nil {
		t.Skipf("audit db not available: %v", err)
	}
	defer pg.Close()

	events := store.NewEvents(pg.Pool())
	ev := audit.Event{
		ID:     "integration-test-" + time.Now().UTC().Format("20060102150405"),
		Source: "integration-test",
		Type:   "test.append_only",
		Time:   time.Now().UTC(),
	}
	if err := events.Insert(ctx, ev); err != nil {
		t.Fatalf("insert as audit_writer: %v", err)
	}

	const deleteQ = `DELETE FROM audit_events WHERE id = $1`
	_, err = pg.Pool().Exec(ctx, deleteQ, ev.ID)
	if err == nil {
		t.Fatal("audit_writer must not be able to DELETE")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Fatalf("expected SQLSTATE 42501, got: %v", err)
	}
}
