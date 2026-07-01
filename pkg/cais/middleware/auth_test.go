package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func devCfg() cais.Config {
	return cais.Config{Env: "development", AdminToken: "secret"}
}

func TestAdminAuth_acceptsBearer(t *testing.T) {
	called := false
	h := AdminAuth(devCfg())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Error("handler not called")
	}
}

func TestAdminAuth_rejectsQueryToken(t *testing.T) {
	h := AdminAuth(devCfg())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin?token=secret", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAdminAuth_rejectsInvalidBearer(t *testing.T) {
	h := AdminAuth(devCfg())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAdminAuth_emptyTokenInDevelopment_passesThrough(t *testing.T) {
	cfg := cais.Config{Env: "development", AdminToken: ""}
	called := false
	h := AdminAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Error("handler not called when ADMIN_TOKEN unset in development")
	}
}

func TestAdminAuth_emptyTokenInProduction_rejects(t *testing.T) {
	cfg := cais.Config{Env: "production", AdminToken: ""}
	h := AdminAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
