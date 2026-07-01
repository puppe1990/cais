package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenAuth_ValidToken(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret")

	called := false
	h := TokenAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin?token=secret", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Error("handler not called")
	}
}

func TestTokenAuth_InvalidToken_Returns401(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret")

	h := TokenAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestTokenAuth_EmptyEnv_PassesThrough(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "")

	called := false
	h := TokenAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Error("handler not called when ADMIN_TOKEN unset")
	}
}
