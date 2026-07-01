package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func TestApp_LogsRoute_AvailableInDevelopment(t *testing.T) {
	a := setupTestAppWithEnv(t, "development")

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Cais Logs") {
		t.Fatalf("body missing logs page: %s", rr.Body.String())
	}
}

func TestApp_LogsRoute_HiddenInProduction(t *testing.T) {
	a := setupTestAppWithEnv(t, "production")

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body: %s", rr.Code, rr.Body.String())
	}
}

func setupTestAppWithEnv(t *testing.T, env string) *App {
	t.Helper()
	root := projectRoot(t)
	catalog := i18n.DefaultCatalog()
	renderer, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), catalog)
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.NewSQLiteStore(":memory:", env)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	cfg := cais.Config{Port: ":0", DBPath: ":memory:", Env: env}
	app, err := New(cfg, Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: filepath.Join(root, "web", "static"),
		Site:      meta.SiteFrom("Cais", ""),
		Catalog:   catalog,
	})
	if err != nil {
		t.Fatal(err)
	}
	return app
}
