package sqlite

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestConfigure_setsWALAndMaxConns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := Configure(db); err != nil {
		t.Fatal(err)
	}

	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want wal", mode)
	}
	if db.Stats().MaxOpenConnections != 1 {
		t.Errorf("MaxOpenConnections = %d, want 1", db.Stats().MaxOpenConnections)
	}
}
