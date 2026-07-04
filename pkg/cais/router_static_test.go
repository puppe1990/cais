package cais

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStaticForEnv_development_noCache(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "development"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	cc := rr.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

func TestStaticForEnv_production_allowsCache(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "production"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Cache-Control") == "no-store" {
		t.Error("production static should not force no-store")
	}
}
