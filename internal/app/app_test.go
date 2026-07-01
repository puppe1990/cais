package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/csrf"
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

func setupTestApp(t *testing.T) *App {
	t.Helper()

	root := projectRoot(t)
	renderer, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"))
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
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

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
	if !strings.Contains(rr.Body.String(), "Bem-vindo") {
		t.Errorf("body missing welcome, got: %s", rr.Body.String())
	}
}

func TestApp_ContactPost_requiresCSRF(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET /contact status = %d", getRR.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=alice@example.com"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF status = %d, want 403", postRR.Code)
	}
}

func TestApp_ContactPost_withCSRF_succeeds(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	var token string
	for _, c := range getRR.Result().Cookies() {
		if c.Name == csrf.CookieName {
			token = c.Value
		}
	}
	if token == "" {
		t.Fatal("missing csrf cookie after GET /contact")
	}

	form := url.Values{}
	form.Set("name", "Alice")
	form.Set("email", "alice@example.com")
	form.Set("csrf_token", token)

	postReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	postReq.Header.Set("HX-Request", "true")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK {
		t.Errorf("POST with CSRF status = %d, want 200, body: %s", postRR.Code, postRR.Body.String())
	}
}
