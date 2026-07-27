package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func TestApp_HealthCheck(t *testing.T) {
	a := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
		t.Errorf("body = %q, want status ok", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"lan_urls"`) {
		t.Errorf("body = %q, want lan_urls", rr.Body.String())
	}
}

func TestApp_HealthCheck_degradedWhenDBClosed(t *testing.T) {
	a := setupTestApp(t)
	_ = a.store.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"status":"degraded"`) {
		t.Errorf("body = %q, want degraded", rr.Body.String())
	}
}

func TestApp_GracefulShutdown(t *testing.T) {
	a := setupTestApp(t)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.RunContext(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("RunContext returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}

func TestApp_HomeRoute(t *testing.T) {
	a := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	// Now served via Inertia; assert protocol markers instead of old template string.
	body := rr.Body.String()
	if !strings.Contains(body, `id="app"`) && !strings.Contains(body, "data-page") {
		t.Errorf("body missing Inertia root markers (home now Inertia), got: %s", body)
	}
}

func TestApp_ProductionBoot(t *testing.T) {
	root := projectRoot(t)
	catalog := i18n.DefaultCatalog()
	renderer, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), catalog)
	if err != nil {
		t.Fatal(err)
	}

	s, err := store.NewSQLiteStore(":memory:", "production")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	cfg := cais.Config{
		Port:       ":0",
		DBPath:     ":memory:",
		Env:        "production",
		AppURL:     "https://example.com",
		AdminToken: "ci-smoke-secret",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("cfg.Validate() = %v", err)
	}

	a, err := New(cfg, Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: filepath.Join(root, "web", "static"),
		Site:      meta.SiteFrom("Cais", cfg.AppURL),
		Catalog:   catalog,
	})
	if err != nil {
		t.Fatal(err)
	}
	h := a.Handler()

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRR := httptest.NewRecorder()
	h.ServeHTTP(healthRR, healthReq)
	if healthRR.Code != http.StatusOK {
		t.Fatalf("GET /health status = %d, want 200", healthRR.Code)
	}
	if !strings.Contains(healthRR.Body.String(), `"status":"ok"`) {
		t.Errorf("health body = %q, want ok", healthRR.Body.String())
	}

	loginReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRR := httptest.NewRecorder()
	h.ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d, want 200", loginRR.Code)
	}
	if loginRR.Header().Get("Strict-Transport-Security") == "" {
		t.Error("missing HSTS header in production")
	}
}

func TestApp_StaticBuildMainJS(t *testing.T) {
	root := projectRoot(t)
	mainJS := filepath.Join(root, "web", "static", "build", "assets", "main.js")
	if _, err := os.Stat(mainJS); os.IsNotExist(err) {
		t.Fatalf("built asset missing at %s — run npm run build first", mainJS)
	}

	a := setupTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/static/build/assets/main.js", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if len(body) < 100 {
		t.Fatalf("body too short: %d bytes", len(body))
	}
	if !strings.Contains(body, "inertia") && !strings.Contains(body, "Inertia") {
		t.Errorf("body missing Inertia reference")
	}
}

// TestApp_Dashboard_Inertia_TDD asserts dashboard X-Inertia response includes totalContacts.
