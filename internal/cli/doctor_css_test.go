package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCSS_missing(t *testing.T) {
	dir := t.TempDir()
	c := checkCSS(dir)
	if c.OK {
		t.Fatal("expected FAIL when styles.css missing")
	}
	if !strings.Contains(c.FixHint, "cais css") {
		t.Errorf("FixHint = %q, want cais css", c.FixHint)
	}
}

func TestCheckCSS_placeholderOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "web/static/css")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	// Scaffold ships a comment-only stub; treating it as OK leaves apps unstyled (#141).
	if err := os.WriteFile(filepath.Join(path, "styles.css"), []byte("/* Run: cais css */\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := checkCSS(dir)
	if c.OK {
		t.Fatal("expected FAIL for placeholder styles.css")
	}
	if !strings.Contains(strings.ToLower(c.Detail), "empty") && !strings.Contains(strings.ToLower(c.Detail), "not built") {
		t.Errorf("Detail = %q, want empty/not built", c.Detail)
	}
}

func TestCheckCSS_emptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "web/static/css")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "styles.css"), []byte("   \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := checkCSS(dir)
	if c.OK {
		t.Fatal("expected FAIL for empty styles.css")
	}
}

func TestCheckCSS_builtOK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "web/static/css")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	// Minified Tailwind output is large and has real selectors.
	body := "*,:after,:before{box-sizing:border-box}.text-stone-900{color:#1c1917}"
	if err := os.WriteFile(filepath.Join(path, "styles.css"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c := checkCSS(dir)
	if !c.OK {
		t.Fatalf("expected OK for built CSS, got %+v", c)
	}
}

func TestStylesCSSReady(t *testing.T) {
	dir := t.TempDir()
	if stylesCSSReady(dir) {
		t.Fatal("missing file should not be ready")
	}
	path := filepath.Join(dir, cssOutput)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("/* Run: cais css */"), 0o644); err != nil {
		t.Fatal(err)
	}
	if stylesCSSReady(dir) {
		t.Fatal("placeholder should not be ready")
	}
	if err := os.WriteFile(path, []byte(".flex{display:flex}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !stylesCSSReady(dir) {
		t.Fatal("built CSS should be ready")
	}
}
