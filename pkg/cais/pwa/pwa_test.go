package pwa

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("Dashboard")
	if cfg.Name != "Dashboard" {
		t.Errorf("Name = %q", cfg.Name)
	}
	if cfg.ThemeColor != ThemeColor {
		t.Errorf("ThemeColor = %q", cfg.ThemeColor)
	}
}

func TestWriteStatic(t *testing.T) {
	dir := t.TempDir()
	if err := WriteStatic(dir, DefaultConfig("My App")); err != nil {
		t.Fatal(err)
	}

	manifest, err := os.ReadFile(filepath.Join(dir, "web/static/manifest.webmanifest"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), "My App") {
		t.Errorf("manifest missing app name: %s", manifest)
	}
	if !strings.Contains(string(manifest), `"display": "fullscreen"`) {
		t.Errorf("manifest should use fullscreen display, got: %s", manifest)
	}

	for _, path := range []string{
		"web/static/js/sw.js",
		"web/static/js/htmx.min.js",
		"web/static/js/cais.js",
		"web/static/offline.html",
		"web/static/icons/icon.png",
		"web/static/img/go-on-cais.jpg",
		"web/static/og.png",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}
}

func TestHeadHTML(t *testing.T) {
	html := HeadHTML()
	if !strings.Contains(html, "manifest.webmanifest") {
		t.Error("HeadHTML missing manifest link")
	}
	if !strings.Contains(html, `apple-mobile-web-app-status-bar-style" content="black-translucent"`) {
		t.Error("HeadHTML should use black-translucent status bar for fullscreen PWA")
	}
}
