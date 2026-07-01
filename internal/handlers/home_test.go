package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matheuspuppe/cais/pkg/cais"
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

func setupTestRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	templatesDir := filepath.Join(projectRoot(t), "web", "templates")
	r, err := cais.NewRendererFromDir(templatesDir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestHomeHandler_Returns200(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "Bem-vindo, Desenvolvedor!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}
