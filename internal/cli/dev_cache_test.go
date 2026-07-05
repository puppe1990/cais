package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaybeBumpDevCache_skipsWithoutSW(t *testing.T) {
	dir := t.TempDir()
	bumped, v, err := maybeBumpDevCache(dir)
	if err != nil {
		t.Fatal(err)
	}
	if bumped {
		t.Error("expected no bump without sw.js")
	}
	if v != 0 {
		t.Errorf("version = %d, want 0", v)
	}
}

func TestMaybeBumpDevCache_increments(t *testing.T) {
	dir := t.TempDir()
	jsDir := filepath.Join(dir, "web", "static", "js")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(jsDir, "sw.js"), []byte("const CACHE_VERSION = 3;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	bumped, v, err := maybeBumpDevCache(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !bumped {
		t.Fatal("expected bump")
	}
	if v != 4 {
		t.Errorf("version = %d, want 4", v)
	}
	data, err := os.ReadFile(filepath.Join(jsDir, "sw.js"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "const CACHE_VERSION = 4;\n" {
		t.Errorf("sw.js = %q", data)
	}
}
