package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenMigratesInitialSchema(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "analysis.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	want := []string{"accounts", "usage_records", "balance_ledger_entries", "sync_cursors", "sync_runs", "upstream_health", "dashboard_users", "web_sessions", "schema_migrations"}
	for _, table := range want {
		var count int
		err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count)
		if err != nil || count != 1 {
			t.Fatalf("table %s missing: count=%d err=%v", table, count, err)
		}
	}
	var foreignKeys int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil || foreignKeys != 1 {
		t.Fatalf("foreign_keys=%d err=%v", foreignKeys, err)
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "analysis.db")
	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	first.Close()
	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var versions int
	if err := second.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&versions); err != nil || versions != 1 {
		t.Fatalf("versions=%d err=%v", versions, err)
	}
}
