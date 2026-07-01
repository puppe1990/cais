package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureRoutesGo = `package app

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, deps.Site, deps.Catalog, cfg)

	r.Get("/", home.ServeHTTP)
	r.Post("/contact", contact.Post)
	r.Group(middleware.AdminAuth(cfg), func(g *cais.Router) {
		g.Get("/admin/items", admin.Index)
		g.Post("/admin/items", admin.Create)
		g.Get("/admin/items/{id}/edit", cais.IntParam("id", admin.Edit))
	})
}
`

func TestParseRoutesContent_detectsRoutes(t *testing.T) {
	entries := parseRoutesContent(fixtureRoutesGo)

	want := []RouteEntry{
		{Method: "GET", Path: "/"},
		{Method: "POST", Path: "/contact"},
		{Method: "GET", Path: "/admin/items"},
		{Method: "POST", Path: "/admin/items"},
		{Method: "GET", Path: "/admin/items/{id}/edit"},
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d routes, want %d: %#v", len(entries), len(want), entries)
	}
	for i, w := range want {
		if entries[i] != w {
			t.Errorf("entry[%d] = %#v, want %#v", i, entries[i], w)
		}
	}
}

func TestFormatRoutes_matchesExpectedOutput(t *testing.T) {
	entries := []RouteEntry{
		{Method: "GET", Path: "/"},
		{Method: "POST", Path: "/contact"},
		{Method: "GET", Path: "/admin/items"},
	}
	got := formatRoutes(entries)
	want := "GET  /\nPOST /contact\nGET  /admin/items"
	if got != want {
		t.Errorf("formatRoutes() = %q, want %q", got, want)
	}
}

func TestParseRoutesFile_readsFixture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "routes.go")
	if err := os.WriteFile(path, []byte(fixtureRoutesGo), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := parseRoutesFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("got %d routes, want 5: %#v", len(entries), entries)
	}
}

func TestCLI_Routes_listsRoutes(t *testing.T) {
	dir := t.TempDir()
	writeRoutesApp(t, dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"routes"}); err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(buf.String())
	want := strings.Join([]string{
		"GET  /",
		"POST /contact",
		"GET  /admin/items",
		"POST /admin/items",
		"GET  /admin/items/{id}/edit",
	}, "\n")
	if out != want {
		t.Errorf("routes output:\n%s\nwant:\n%s", out, want)
	}
}

func TestCLI_Routes_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"routes"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func writeRoutesApp(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"go.mod": `module testapp

require github.com/puppe1990/cais v0.3.0
`,
		"internal/app/routes.go": fixtureRoutesGo,
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
