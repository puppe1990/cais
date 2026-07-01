package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListSeeds_findsSeedCalls(t *testing.T) {
	dir := t.TempDir()
	content := `package db

import "example.com/internal/store"

func RunSeeds(s store.Store) error {
	// cais:seeds
	if err := s.SeedDemoBookmarks(); err != nil {
		return err
	}
	return nil
}
`
	writeFile(t, dir, "internal/db/seeds.go", content)

	items, err := listSeedsInDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0] != "SeedDemoBookmarks" {
		t.Fatalf("items = %v", items)
	}
}

func TestCLI_DBSeed_listFlag(t *testing.T) {
	dir := t.TempDir()
	writeMinimalAppWithSeeds(t, dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := c.Run([]string{"db", "seed", "--list"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "RunSeeds") {
		t.Errorf("output = %q", buf.String())
	}
}

func writeMinimalAppWithSeeds(t *testing.T, dir string) {
	t.Helper()
	writeMinimalApp(t, dir)
	seedsDir := filepath.Join(dir, "internal/db")
	if err := os.MkdirAll(seedsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedsDir, "seeds.go"), []byte(`package db

import "testapp/internal/store"

func RunSeeds(s store.Store) error {
	return nil
}
`), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
