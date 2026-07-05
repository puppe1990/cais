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
	if cfg.Display != "fullscreen" {
		t.Errorf("Display = %q, want fullscreen", cfg.Display)
	}
}

func TestCaisJS_hasSSEReconnect(t *testing.T) {
	data, err := assets.ReadFile("assets/cais.js")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"htmx:sseClose", "data-cais-sse-persist", "reconnectChatSSE"} {
		if !strings.Contains(content, want) {
			t.Errorf("cais.js missing %q", want)
		}
	}
}

func TestCaisJS_hasChatAgentModule(t *testing.T) {
	data, err := assets.ReadFile("assets/cais.js")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		"data-cais-chat",
		"finalizeChatStream",
		"formatMessageTimes",
		"cais-msg-time",
		"data-cais-live",
		"chatStickToBottom",
		"chat-scroll-down",
		"window.caisFinalizeChatStream",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("cais.js missing chat agent helper %q", want)
		}
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
		"web/static/js/sse-ext.min.js",
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

func TestWriteStatic_customDisplay(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig("App")
	cfg.Display = "standalone"
	if err := WriteStatic(dir, cfg); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(dir, "web/static/manifest.webmanifest"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), `"display": "standalone"`) {
		t.Errorf("manifest should respect Display, got: %s", manifest)
	}
}

func TestWriteStatic_customIcon(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "brand")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Copy embedded default to a distinct source file.
	fsys, err := FS()
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(fsys, "icon.png")
	if err != nil {
		t.Fatal(err)
	}
	iconPath := filepath.Join(srcDir, "logo.png")
	if err := os.WriteFile(iconPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig("Branded")
	cfg.IconPath = iconPath
	if err := WriteStatic(dir, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "web/static/icons/icon-192.png"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected generated icon")
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
