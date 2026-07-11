package store

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/001_initial.sql
var initialMigration string

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	var exists int
	if err := db.QueryRow(`SELECT count(*) FROM schema_migrations WHERE version=1`).Scan(&exists); err != nil {
		return fmt.Errorf("read migration version: %w", err)
	}
	if exists == 1 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(initialMigration); err != nil {
		return fmt.Errorf("apply migration 1: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(1, strftime('%Y-%m-%dT%H:%M:%fZ','now'))`); err != nil {
		return fmt.Errorf("record migration 1: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration 1: %w", err)
	}
	return nil
}
