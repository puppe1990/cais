package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetGoModReplace_addsReplace(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/acme/app\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := setGoModReplace(dir, "../Cais"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "replace github.com/puppe1990/cais => ../Cais") {
		t.Errorf("go.mod missing replace: %s", body)
	}
}

func TestSetGoModReplace_updatesExisting(t *testing.T) {
	dir := t.TempDir()
	initial := "module github.com/acme/app\n\ngo 1.26\n\nreplace github.com/puppe1990/cais => ../old\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := setGoModReplace(dir, "../Cais"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if strings.Contains(s, "../old") {
		t.Errorf("go.mod should update replace path: %s", s)
	}
	if !strings.Contains(s, "replace github.com/puppe1990/cais => ../Cais") {
		t.Errorf("go.mod missing updated replace: %s", s)
	}
}

func TestRemoveGoModReplace(t *testing.T) {
	dir := t.TempDir()
	initial := "module github.com/acme/app\n\ngo 1.26\n\nreplace github.com/puppe1990/cais => ../Cais\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := removeGoModReplace(dir); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "replace github.com/puppe1990/cais") {
		t.Errorf("go.mod should not contain replace: %s", body)
	}
}
