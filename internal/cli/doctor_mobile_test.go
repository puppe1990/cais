package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_MobileOKOnFreshScaffold(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "mobile",
		ModulePath: "github.com/puppe1990/mobile",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor --mobile failed: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[ok] flash template") {
		t.Errorf("expected flash template ok, got:\n%s", out)
	}
	if !strings.Contains(out, "[ok] CSP fonts") {
		t.Errorf("expected CSP fonts ok, got:\n%s", out)
	}
}

func TestDoctor_MobileWarnsGoogleFonts(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "mobile",
		ModulePath: "github.com/puppe1990/mobile",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	cssPath := filepath.Join(dir, "input.css")
	body, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatal(err)
	}
	patched := "@import url('https://fonts.googleapis.com/css');\n" + string(body)
	if err := os.WriteFile(cssPath, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor should pass with warning: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "[warn] CSP fonts") {
		t.Errorf("expected CSP fonts warning, got:\n%s", buf.String())
	}
}
