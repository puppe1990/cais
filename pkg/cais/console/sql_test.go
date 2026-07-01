package console

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestFormatSQLRows_withData(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`CREATE TABLE items (id INTEGER, name TEXT); INSERT INTO items VALUES (1, 'alpha'), (2, 'beta')`); err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT id, name FROM items ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rows.Close() })

	out, err := formatSQLRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"id", "name", "1", "alpha", "2", "beta"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q, got:\n%s", want, out)
		}
	}
}

func TestFormatSQLRows_emptyResult(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`CREATE TABLE empty (id INTEGER)`); err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT id FROM empty")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rows.Close() })

	out, err := formatSQLRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	if out != "(0 rows)" {
		t.Fatalf("output = %q, want %q", out, "(0 rows)")
	}
}
