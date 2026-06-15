// Command migrate applies SQL migrations for Go services that own schemas.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
)

func auditDatabase() string {
	if d := os.Getenv("AUDIT_DB"); d != "" {
		return d
	}
	return "audit_db"
}

var services = []struct {
	name string
	db   string
}{
	{name: "user-service", db: ""},
	{name: "auth-service", db: ""},
	{name: "file-service", db: ""},
	{name: "audit-service", db: auditDatabase()},
	{name: "worker-service", db: ""},
}

func main() {
	config.LoadEnvFile()
	ctx := context.Background()

	for _, svc := range services {
		cfg := db.ConfigFromEnv(svc.db)
		if config.IsPlaceholder(cfg.Host) {
			log.Printf("skip %s: postgres host not configured", svc.name)
			continue
		}
		pg, err := db.Open(ctx, cfg)
		if err != nil {
			log.Fatalf("%s: connect: %v", svc.name, err)
		}
		if err := db.RunMigrations(cfg, svc.name); err != nil {
			pg.Close()
			log.Fatalf("%s: migrate: %v", svc.name, err)
		}
		pg.Close()
		fmt.Printf("migrations applied: %s\n", svc.name)
	}
}
