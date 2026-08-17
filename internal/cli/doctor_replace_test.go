package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckLocalCaisReplace_nilWhenNoReplace(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module app\n\nrequire github.com/puppe1990/cais v0.8.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if c := checkLocalCaisReplace(dir); c != nil {
		t.Fatalf("expected nil without replace, got %+v", c)
	}
}

func TestCheckLocalCaisReplace_warnsOutsideCI(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	dir := t.TempDir()
	mod := "module app\n\nrequire github.com/puppe1990/cais v0.8.1\n\nreplace github.com/puppe1990/cais => ../cais\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}

	c := checkLocalCaisReplace(dir)
	if c == nil {
		t.Fatal("expected warn when replace is present")
	}
	if c.OK || !c.Optional {
		t.Fatalf("want optional warn, got %+v", c)
	}
	if c.FixHint == "" {
		t.Fatal("expected unlink fix hint")
	}
}
