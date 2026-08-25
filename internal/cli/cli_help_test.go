package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Help(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "cais new") {
		t.Error("help missing cais new")
	}
}

func TestCLI_Help_IncludesResource(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "[--dry-run] resource") {
		t.Error("help missing g resource")
	}
}

func TestCLI_Help_IncludesModuleFlag(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "--module") {
		t.Error("help missing --module flag")
	}
}

func TestCLI_NewHelp_doesNotCreateDir(t *testing.T) {
	assertNewHelpDoesNotScaffold(t, []string{"new", "--help"}, "--help")
}

func TestCLI_NewHelp_shortFlag(t *testing.T) {
	assertNewHelpDoesNotScaffold(t, []string{"new", "-h"}, "-h")
}

func TestCLI_New_helpWithName_doesNotScaffold(t *testing.T) {
	assertNewHelpDoesNotScaffold(t, []string{"new", "myapp", "--help"}, "myapp", "--help")
}

func assertNewHelpDoesNotScaffold(t *testing.T, args []string, forbiddenDirs ...string) {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	t.Chdir(dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run(args); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "cais new") {
		t.Errorf("help output missing cais new usage: %q", out)
	}
	if strings.Contains(out, "Created app") {
		t.Errorf("help must not scaffold: %q", out)
	}
	for _, name := range forbiddenDirs {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Errorf("cais %v must not create directory %q", args, name)
		}
	}
}
