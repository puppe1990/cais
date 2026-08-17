package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestCheckLocalCaisReplace_failsInCI(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("GITHUB_ACTIONS", "")
	dir := t.TempDir()
	mod := "module app\n\nreplace github.com/puppe1990/cais => ../cais\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}

	c := checkLocalCaisReplace(dir)
	if c == nil || c.OK || c.Optional {
		t.Fatalf("want hard FAIL in CI, got %+v", c)
	}
}

func TestCheckLocalCaisReplace_failsOnGitHubActions(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "true")
	dir := t.TempDir()
	mod := "module app\n\nreplace github.com/puppe1990/cais => ../cais\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}

	c := checkLocalCaisReplace(dir)
	if c == nil || c.OK || c.Optional {
		t.Fatalf("want hard FAIL on GITHUB_ACTIONS, got %+v", c)
	}
}

func TestRunDoctor_caisReplaceWarnsLocally(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)
	if err := setGoModReplace(dir, "../cais"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("local doctor should warn, not fail: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[warn] cais replace") {
		t.Fatalf("expected warn, got:\n%s", out)
	}
	if !strings.Contains(out, "cais link --unlink") {
		t.Fatalf("expected unlink hint, got:\n%s", out)
	}
}

func TestRunDoctor_caisReplaceFailsInCI(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)
	if err := setGoModReplace(dir, "../cais"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err == nil {
		t.Fatalf("CI doctor should fail on replace, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "[FAIL] cais replace") {
		t.Fatalf("expected FAIL, got:\n%s", buf.String())
	}
}
