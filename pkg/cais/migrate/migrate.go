package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
)

const schemaTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY NOT NULL,
  applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);`

// Entry describes a migration file and whether it has been applied.
type Entry struct {
	Version string
	Applied bool
}

// ApplyDir runs pending SQL migrations from a filesystem directory.
func ApplyDir(db *sql.DB, dir string) error {
	return Apply(db, os.DirFS(dir), ".")
}

// StatusDir returns migration status for SQL files in a filesystem directory.
func StatusDir(db *sql.DB, dir string) ([]Entry, error) {
	return Status(db, os.DirFS(dir), ".")
}

// RollbackLastDir removes the last applied migration record from a filesystem directory.
func RollbackLastDir(db *sql.DB, dir string) (string, error) {
	return RollbackLast(db, os.DirFS(dir), ".")
}

// Apply runs pending SQL migrations from dir inside migrations in sorted order.
func Apply(db *sql.DB, migrations fs.FS, dir string) error {
	if _, err := db.Exec(schemaTable); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	files, err := listSQL(migrations, dir)
	if err != nil {
		return err
	}

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		applied, err := isApplied(db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlPath := path.Join(dir, name)
		sqlBytes, err := fs.ReadFile(migrations, sqlPath)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}
	}

	return nil
}

// Status returns migration files in order with applied state.
func Status(db *sql.DB, migrations fs.FS, dir string) ([]Entry, error) {
	if _, err := db.Exec(schemaTable); err != nil {
		return nil, fmt.Errorf("ensure schema_migrations: %w", err)
	}

	files, err := listSQL(migrations, dir)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		applied, err := isApplied(db, version)
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{Version: version, Applied: applied})
	}
	return entries, nil
}

// RollbackLast removes the last applied migration record from schema_migrations.
// It does not execute SQL down migrations.
func RollbackLast(db *sql.DB, migrations fs.FS, dir string) (string, error) {
	entries, err := Status(db, migrations, dir)
	if err != nil {
		return "", err
	}

	var lastApplied string
	for _, e := range entries {
		if e.Applied {
			lastApplied = e.Version
		}
	}
	if lastApplied == "" {
		return "", fmt.Errorf("no applied migrations to roll back")
	}

	if _, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", lastApplied); err != nil {
		return "", fmt.Errorf("remove migration %s: %w", lastApplied, err)
	}

	return lastApplied, nil
}

func listSQL(migrations fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(migrations, dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func isApplied(db *sql.DB, version string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return count > 0, nil
}
