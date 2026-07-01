package cais

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRouter_GetRoute(t *testing.T) {
	r := NewRouter()
	called := false
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRouter_PostRoute(t *testing.T) {
	r := NewRouter()
	r.Post("/submit", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
}

func TestRouter_StaticFiles(t *testing.T) {
	dir := t.TempDir()
	cssDir := filepath.Join(dir, "css")
	if err := os.MkdirAll(cssDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cssDir, "styles.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.Static("/static", dir)

	req := httptest.NewRequest(http.MethodGet, "/static/css/styles.css", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); body != "body{}" {
		t.Errorf("body = %q, want %q", body, "body{}")
	}
}

func TestRouter_Middleware(t *testing.T) {
	r := NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Test", "ok")
			next.ServeHTTP(w, req)
		})
	})
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Header().Get("X-Test") != "ok" {
		t.Errorf("X-Test = %q, want %q", rr.Header().Get("X-Test"), "ok")
	}
}