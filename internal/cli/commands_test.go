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
	for _, cmd := range []string{"cais install", "cais css", "cais dev", "cais build", "cais server", "cais db migrate", "cais db status", "cais db rollback", "cais db prune-sessions", "cais routes"} {
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
