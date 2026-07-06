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
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg, nil)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, deps.Site, deps.Catalog, cfg, nil)

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
		{Method: "GET", Path: "/", Handler: "home.ServeHTTP"},
		{Method: "POST", Path: "/contact", Handler: "contact.Post"},
		{Method: "GET", Path: "/admin/items", Handler: "admin.Index", Middleware: "middleware.AdminAuth(cfg)"},
		{Method: "POST", Path: "/admin/items", Handler: "admin.Create", Middleware: "middleware.AdminAuth(cfg)"},
		{Method: "GET", Path: "/admin/items/{id}/edit", Handler: "cais.IntParam(\"id\", admin.Edit)", Middleware: "middleware.AdminAuth(cfg)"},
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

func TestParseRoutesContent_verboseIncludesMiddleware(t *testing.T) {
	entries := parseRoutesVerbose(fixtureRoutesGo)
	if len(entries) != 5 {
		t.Fatalf("got %d routes, want 5", len(entries))
	}
	var adminRoute *RouteEntry
	for i := range entries {
		if entries[i].Path == "/admin/items" && entries[i].Method == "GET" {
			adminRoute = &entries[i]
			break
		}
	}
	if adminRoute == nil {
		t.Fatal("missing admin items route")
	}
	if adminRoute.Middleware != "middleware.AdminAuth(cfg)" {
		t.Errorf("Middleware = %q", adminRoute.Middleware)
	}
	if adminRoute.Handler != "admin.Index" {
		t.Errorf("Handler = %q", adminRoute.Handler)
	}
}

func TestCLI_Routes_verboseFlag(t *testing.T) {
	dir := t.TempDir()
	writeRoutesApp(t, dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := c.Run([]string{"routes", "--verbose"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "admin.Index") {
		t.Errorf("verbose output missing handler: %q", out)
	}
	if !strings.Contains(out, "AdminAuth") {
		t.Errorf("verbose output missing middleware: %q", out)
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
