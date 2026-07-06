package cais

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeDevTemplates(t *testing.T, dir string, homeBody string) {
	t.Helper()
	for _, sub := range []string{"layouts", "pages", "partials"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"layouts/base.html": `{{ define "base" }}<html><body>{{ template "content" . }}</body></html>{{ end }}`,
		"pages/home.html":   `{{ define "content" }}` + homeBody + `{{ end }}`,
	}
	for path, body := range files {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRenderer_ReloadFromDir_picksUpChanges(t *testing.T) {
	dir := t.TempDir()
	writeDevTemplates(t, dir, `<p>v1</p>`)

	r, err := NewRendererFromDir(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	if err := r.Render(&buf, "base", "home", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "v1") {
		t.Fatalf("first render: %s", buf.String())
	}

	writeDevTemplates(t, dir, `<p>v2</p>`)
	if err := r.ReloadFromDir(dir); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := r.Render(&buf, "base", "home", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "v2") {
		t.Fatalf("after reload: %s", buf.String())
	}
}

func TestNewRendererForEnv_developmentUsesDisk(t *testing.T) {
	dir := t.TempDir()
	writeDevTemplates(t, dir, `<p>disk</p>`)

	cfg := Config{Env: "development"}
	r, err := NewRendererForEnv(cfg, nil, dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	if err := r.Render(&buf, "base", "home", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "disk") {
		t.Fatalf("body = %s", buf.String())
	}
}

func TestNewRendererForEnv_stagingUsesDisk(t *testing.T) {
	dir := t.TempDir()
	writeDevTemplates(t, dir, `<p>staging-disk</p>`)

	cfg := Config{Env: "staging"}
	r, err := NewRendererForEnv(cfg, nil, dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	if err := r.Render(&buf, "base", "home", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "staging-disk") {
		t.Fatalf("body = %s", buf.String())
	}
}

func TestNewRendererForEnv_inertiaOnlyUsesStub(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.html"), []byte("<html>{{ .inertia }}</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Env: "development"}
	r, err := NewRendererForEnv(cfg, nil, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.pages) != 0 {
		t.Fatal("expected stub renderer with no pages")
	}
}

func TestNewRendererForEnv_productionUsesEmbed(t *testing.T) {
	tmplFS, err := fsSubTestTemplates()
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{Env: "production"}
	r, err := NewRendererForEnv(cfg, tmplFS, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	if err := r.Render(&buf, "base", "home", map[string]string{"Name": "World"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Hello, World!") {
		t.Fatalf("body = %s", buf.String())
	}
}

func TestStartTemplateWatcher_reloadOnChange(t *testing.T) {
	dir := t.TempDir()
	writeDevTemplates(t, dir, `<p>before</p>`)

	r, err := NewRendererFromDir(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	StartTemplateWatcher(dir, r)
	t.Cleanup(func() { stopTemplateWatcher() })

	writeDevTemplates(t, dir, `<p>after</p>`)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var buf strings.Builder
		_ = r.Render(&buf, "base", "home", nil)
		if strings.Contains(buf.String(), "after") {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("watcher did not reload templates in time")
}

func fsSubTestTemplates() (fs.FS, error) {
	return fs.Sub(testTemplates, "testdata/templates")
}
