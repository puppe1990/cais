package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_AllOK(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir); err != nil {
		t.Fatalf("doctor failed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "htmx.min.js") {
		t.Error("missing htmx check")
	}
}

func TestDoctor_AirOptionalWhenMissing(t *testing.T) {
	if _, err := exec.LookPath("air"); err == nil {
		t.Skip("air installed; optional-missing path not exercised")
	}
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir); err != nil {
		t.Fatalf("doctor should pass without air: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "[warn] air") {
		t.Errorf("expected air warning, got:\n%s", buf.String())
	}
}

func TestDoctor_QualityToolingWarnsWhenMissing(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "legacy",
		ModulePath: "github.com/puppe1990/legacy",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, ".github/workflows/ci.yml")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir); err != nil {
		t.Fatalf("doctor should pass with optional warning: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[warn] quality tooling") {
		t.Errorf("expected quality tooling warning, got:\n%s", out)
	}
	if !strings.Contains(out, "cais g ci") {
		t.Errorf("expected fix hint, got:\n%s", out)
	}
}
