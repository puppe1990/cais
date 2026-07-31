package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Help_IncludesAppCommands(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	for _, cmd := range []string{"cais install", "cais css", "cais dev", "cais build", "cais server", "cais db migrate", "cais db status", "cais db rollback", "cais db prune-sessions", "cais db seed", "cais routes", "cais version", "cais g [--dry-run] ci", "cais g [--dry-run] console", "cais destroy"} {
		if !strings.Contains(buf.String(), cmd) {
			t.Errorf("help missing %q", cmd)
		}
	}
}

func TestCLI_Install_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"install"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestCLI_CSS_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"css"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestCLI_Dev_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"dev"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestCLI_Build_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"build"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestFindAir(t *testing.T) {
	// always returns empty or a path — must not panic
	_ = findAir()
}

func TestRunTailwindBuild_missingInput(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\nrequire github.com/puppe1990/cais v0.3.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runTailwindBuild(dir, false)
	if err == nil {
		t.Fatal("expected error without input.css")
	}
}

func TestEnsureStylesCSS_skipsWhenReady(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, cssOutput)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(".p-4{padding:1rem}"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := ensureStylesCSS(&buf, dir); err != nil {
		t.Fatalf("ready CSS should not error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no build log when ready, got %q", buf.String())
	}
}

func TestEnsureStylesCSS_errorsWithoutInput(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	err := ensureStylesCSS(&buf, dir)
	if err == nil {
		t.Fatal("expected error when styles missing and no input.css")
	}
	if !strings.Contains(err.Error(), "unstyled") && !strings.Contains(err.Error(), cssOutput) {
		t.Errorf("error = %v", err)
	}
}
