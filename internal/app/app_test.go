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
	if !strings.Contains(rr.Body.String(), "on Cais!") {
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

func TestApp_LoginPost_requiresCSRF(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=demo@example.com&password=password"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF status = %d, want 403", postRR.Code)
	}
}

func TestApp_LoginPost_withCSRF_redirects(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	var token string
	for _, c := range getRR.Result().Cookies() {
		if c.Name == csrf.CookieName {
			token = c.Value
		}
	}
	if token == "" {
		t.Fatal("missing csrf cookie after GET /login")
	}

	form := url.Values{}
	form.Set("email", "demo@example.com")
	form.Set("password", "password")
	form.Set("csrf_token", token)

	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusSeeOther {
		t.Errorf("POST with CSRF status = %d, want 303, body: %s", postRR.Code, postRR.Body.String())
	}
}

func TestApp_Dashboard_requiresAuth(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rr.Header().Get("Location"))
	}
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

func TestApp_AuthFlow_loginDashboardLogout(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	getLogin := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRR := httptest.NewRecorder()
	h.ServeHTTP(loginRR, getLogin)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d", loginRR.Code)
	}
	csrfToken := csrfTokenFromResponse(t, loginRR.Result())

	form := url.Values{}
	form.Set("email", "demo@example.com")
	form.Set("password", "password")
	form.Set("csrf_token", csrfToken)

	postLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postLogin.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	loginPostRR := httptest.NewRecorder()
	h.ServeHTTP(loginPostRR, postLogin)
	if loginPostRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /login status = %d, want 303", loginPostRR.Code)
	}
	sessionCookie := sessionCookieFromResponse(t, loginPostRR.Result())
	if sessionCookie == nil {
		t.Fatal("missing session cookie after login")
	}

	dashReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	for _, c := range loginPostRR.Result().Cookies() {
		dashReq.AddCookie(c)
	}
	dashRR := httptest.NewRecorder()
	h.ServeHTTP(dashRR, dashReq)
	if dashRR.Code != http.StatusOK {
		t.Fatalf("GET /dashboard status = %d, want 200", dashRR.Code)
	}
	if !strings.Contains(dashRR.Body.String(), "Welcome!") {
		t.Errorf("dashboard missing login flash, body: %s", dashRR.Body.String())
	}

	logoutForm := url.Values{}
	logoutForm.Set("csrf_token", csrfToken)
	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(logoutForm.Encode()))
	logoutReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	logoutReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	logoutReq.AddCookie(sessionCookie)
	logoutRR := httptest.NewRecorder()
	h.ServeHTTP(logoutRR, logoutReq)
	if logoutRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /logout status = %d, want 303", logoutRR.Code)
	}
	if logoutRR.Header().Get("Location") != "/login" {
		t.Errorf("logout Location = %q, want /login", logoutRR.Header().Get("Location"))
	}

	dashAfterLogout := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashAfterLogout.AddCookie(sessionCookie)
	dashAfterRR := httptest.NewRecorder()
	h.ServeHTTP(dashAfterRR, dashAfterLogout)
	if dashAfterRR.Code != http.StatusSeeOther {
		t.Errorf("GET /dashboard after logout status = %d, want 303", dashAfterRR.Code)
	}
}

func TestApp_ContactPost_validationWithCSRF_returns422(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	csrfToken := csrfTokenFromResponse(t, getRR.Result())

	form := url.Values{}
	form.Set("name", "")
	form.Set("email", "alice@example.com")
	form.Set("csrf_token", csrfToken)

	postReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	postReq.Header.Set("HX-Request", "true")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusUnprocessableEntity {
		t.Errorf("POST status = %d, want 422, body: %s", postRR.Code, postRR.Body.String())
	}
	if !strings.Contains(postRR.Body.String(), "Name is required") {
		t.Errorf("body missing validation error: %s", postRR.Body.String())
	}
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
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}
