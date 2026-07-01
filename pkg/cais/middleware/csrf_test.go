package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/csrf"
)

func TestCSRF_safeMethod_setsCookie(t *testing.T) {
	called := false
	h := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if csrf.TokenFromRequest(r) == "" {
			t.Error("expected token in request context")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	found := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == csrf.CookieName {
			found = true
		}
	}
	if !found {
		t.Error("csrf cookie not set on GET")
	}
}

func TestCSRF_unsafeMethod_rejectsMissingToken(t *testing.T) {
	h := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=a"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

func TestCSRF_unsafeMethod_acceptsValidToken(t *testing.T) {
	called := false
	h := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	body := "name=a&csrf_token=secret"
	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "secret"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestCSRF_skipsHealthAndStatic(t *testing.T) {
	called := false
	h := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler not called for /health")
	}
}
