package pwa

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBumpCacheVersion_incrementsSW(t *testing.T) {
	dir := t.TempDir()
	jsDir := filepath.Join(dir, "web/static/js")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const initial = `const CACHE_VERSION = 2;
const CACHE = "cais-static-v" + CACHE_VERSION;
`
	if err := os.WriteFile(filepath.Join(jsDir, "sw.js"), []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	v, err := BumpCacheVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v != 3 {
		t.Errorf("version = %d, want 3", v)
	}

	body, err := os.ReadFile(filepath.Join(jsDir, "sw.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "const CACHE_VERSION = 3;") {
		t.Errorf("sw.js not bumped: %s", body)
	}
}

func TestWriteStatic_includesCacheVersion(t *testing.T) {
	dir := t.TempDir()
	if err := WriteStatic(dir, DefaultConfig("Test")); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "web/static/js/sw.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "CACHE_VERSION") {
		t.Errorf("sw.js missing CACHE_VERSION: %s", body)
	}
}
