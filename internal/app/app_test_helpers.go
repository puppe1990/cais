package app

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/csrf"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func setupTestInertiaFromTemplates(t *testing.T) *inertia.Inertia {
	t.Helper()
	root := projectRoot(t)
	i, err := inertia.NewFromFile(filepath.Join(root, "web", "templates", "app.html"))
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func setupTestApp(t *testing.T) *App {
	t.Helper()

	root := projectRoot(t)
	catalog := i18n.DefaultCatalog()
	renderer, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), catalog)
	if err != nil {
		t.Fatal(err)
	}

	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	cfg := cais.Config{Port: ":0", DBPath: ":memory:", Env: "test"}
	a, err := New(cfg, Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: filepath.Join(root, "web", "static"),
		Site:      meta.SiteFrom("Cais", ""),
		Catalog:   catalog,
		Inertia:   setupTestInertiaFromTemplates(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func csrfTokenFromResponse(t *testing.T, res *http.Response) string {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == csrf.CookieName {
			return c.Value
		}
	}
	t.Fatal("missing csrf cookie")
	return ""
}

func sessionCookieFromResponse(t *testing.T, res *http.Response) *http.Cookie {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == "cais_session" {
			return c
		}
	}
	return nil
}

func setupTestAppDev(t *testing.T) *App {
	t.Helper()

	root := projectRoot(t)
	catalog := i18n.DefaultCatalog()
	renderer, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), catalog)
	if err != nil {
		t.Fatal(err)
	}

	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	cfg := cais.Config{Port: ":0", DBPath: ":memory:", Env: "development"}
	a, err := New(cfg, Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: filepath.Join(root, "web", "static"),
		Site:      meta.SiteFrom("Cais", ""),
		Catalog:   catalog,
		Inertia:   setupTestInertiaFromTemplates(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// TestApp_HomeRoute_Inertia asserts the real production home route entry point
// returns Inertia protocol responses (root HTML shell or JSON for X-Inertia).
// Written first per TDD; will fail until gonertia wired in home path.
