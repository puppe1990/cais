package cais

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveWebDir_override(t *testing.T) {
	dir := t.TempDir()
	static := filepath.Join(dir, "custom", "static")
	if err := os.MkdirAll(static, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveWebDir("static", static)
	if err != nil {
		t.Fatal(err)
	}
	if got != static {
		t.Errorf("got %q, want %q", got, static)
	}
}

func TestResolveWebDir_overrideMissing(t *testing.T) {
	_, err := ResolveWebDir("static", "/nonexistent/static")
	if err == nil {
		t.Fatal("expected error for missing override")
	}
	if got := err.Error(); !strings.Contains(got, "STATIC_DIR") || !strings.Contains(got, "WorkingDirectory") {
		t.Errorf("error should mention deploy hints, got: %s", got)
	}
}

func TestResolveWebDir_walkFromCwd(t *testing.T) {
	root := t.TempDir()
	static := filepath.Join(root, "web", "static")
	if err := os.MkdirAll(static, 0o755); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "cmd", "server")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)

	got, err := ResolveWebDir("static", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != static {
		t.Errorf("got %q, want %q", got, static)
	}
}
