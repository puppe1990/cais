package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_DBStatus_listsMigrations(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"db", "status"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "001_contacts") {
		t.Errorf("status output missing migration: %q", buf.String())
	}
}

func TestCLI_DBRollback_removesLastMigration(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	c := &CLI{Out: &bytes.Buffer{}}
	if err := c.Run([]string{"db", "migrate"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c2 := &CLI{Out: &buf}
	if err := c2.Run([]string{"db", "rollback"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "001_contacts") {
		t.Errorf("rollback output missing version: %q", out)
	}
	if !strings.Contains(out, "does not run SQL down migrations") {
		t.Errorf("rollback output missing down-migration notice: %q", out)
	}
}

func TestCLI_DBMigrate_isIdempotent(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	c := &CLI{Out: &bytes.Buffer{}}
	if err := c.Run([]string{"db", "migrate"}); err != nil {
		t.Fatal(err)
	}
	if err := c.Run([]string{"db", "migrate"}); err != nil {
		t.Fatalf("second migrate failed: %v", err)
	}
}

func writeMinimalApp(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"go.mod": `module testapp

require github.com/puppe1990/cais v0.3.0
`,
		"internal/store/migrations/001_contacts.sql": `CREATE TABLE IF NOT EXISTS contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL
);`,
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
