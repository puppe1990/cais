package session

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := EnsureSQLiteSchema(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSQLiteStore_CreateGetDelete(t *testing.T) {
	store := NewSQLiteStore(testDB(t))

	token, err := store.Create(42)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected token")
	}

	id, ok := store.Get(token)
	if !ok || id != 42 {
		t.Fatalf("Get = (%d, %v), want (42, true)", id, ok)
	}

	store.Delete(token)
	if _, ok := store.Get(token); ok {
		t.Fatal("session should be deleted")
	}
}
