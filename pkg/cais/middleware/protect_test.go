package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtect_NoTokenEnv_PassesThrough(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "")

	called := false
	h := Protect(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	h(rr, req)

	if !called {
		t.Error("handler not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestProtect_WithToken_RejectsUnauthorized(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret")

	h := Protect(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestProtect_WithToken_AcceptsBearer(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret")

	called := false
	h := Protect(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	h(rr, req)

	if !called {
		t.Error("handler not called")
	}
}
