package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies SQL migrations from migrations/<service>/ relative to the
// process working directory (go/ when using make migrate or cd go && go run ./cmd/migrate).
func RunMigrations(cfg Config, service string) error {
	dir, err := filepath.Abs(filepath.Join("migrations", service))
	if err != nil {
		return err
	}
	sourceURL := "file://" + filepath.ToSlash(dir)
	m, err := migrate.New(sourceURL, cfg.MigrateDSN())
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
