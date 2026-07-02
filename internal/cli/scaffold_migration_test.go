package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldMigration_usesMaxMigrationNumber(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "internal", "store", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001_contacts.sql", "003_other.sql"} {
		if err := os.WriteFile(filepath.Join(migrationsDir, name), []byte("-- up\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := scaffoldMigration(dir, "posts", false); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(migrationsDir, "004_posts.sql")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected %s: %v", want, err)
	}
	if _, err := os.Stat(filepath.Join(migrationsDir, "003_posts.sql")); err == nil {
		t.Fatal("should not create 003_posts.sql when 003 is taken")
	}
}

func TestScaffoldMigration_numbersAfterSQLOnly(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "internal", "store", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_contacts.sql"), []byte("CREATE TABLE contacts (id INTEGER PRIMARY KEY);"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(migrationsDir, ".gitkeep"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldMigration(dir, "posts", false); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(migrationsDir, "002_posts.sql")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected %s: %v", want, err)
	}
}
