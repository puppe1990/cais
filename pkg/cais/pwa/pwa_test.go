package pwa

import (
	"io/fs"
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
	if !strings.Contains(string(manifest), `"display": "standalone"`) {
		t.Errorf("manifest should use standalone display, got: %s", manifest)
	}
	if !strings.Contains(string(manifest), "icon-192.png") {
		t.Errorf("manifest missing 192 icon: %s", manifest)
	}
	if !strings.Contains(string(manifest), "icon-512.png") {
		t.Errorf("manifest missing 512 icon: %s", manifest)
	}

	for _, path := range []string{
		"web/static/js/sw.js",
		"web/static/js/htmx.min.js",
		"web/static/js/idiomorph-ext.min.js",
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

func TestRegisterScriptForEnv_developmentClearsSW(t *testing.T) {
	script := RegisterScriptForEnv("development")
	if !strings.Contains(script, "unregister") {
		t.Errorf("dev script should unregister SW: %s", script)
	}
	if strings.Contains(script, ".register(") {
		t.Errorf("dev script should not register SW: %s", script)
	}
}

func TestRegisterScriptForEnv_productionRegistersSW(t *testing.T) {
	script := RegisterScriptForEnv("production")
	if !strings.Contains(script, "register(") {
		t.Errorf("prod script should register SW: %s", script)
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

func TestFS(t *testing.T) {
	fsys, err := FS()
	if err != nil {
		t.Fatalf("FS() error = %v", err)
	}
	if _, err := fs.Stat(fsys, "sw.js"); err != nil {
		t.Fatalf("FS() missing sw.js: %v", err)
	}
}
