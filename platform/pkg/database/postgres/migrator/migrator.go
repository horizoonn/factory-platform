package migrator

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

type Migrator struct {
	db            *sql.DB
	migrationsDir string
}

func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

func (m *Migrator) Up(ctx context.Context) error {
	err := goose.UpContext(ctx, m.db, m.migrationsDir)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
