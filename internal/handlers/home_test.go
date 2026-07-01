package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func testSite() meta.Site {
	return meta.Site{AppName: "Cais", AppURL: "https://cais.example.com"}
}

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
	r, err := cais.NewRendererFromDir(templatesDir, i18n.DefaultCatalog())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestHomeHandler_Returns200(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), i18n.DefaultCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), i18n.DefaultCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "on Cais!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "go-on-cais.jpg") {
		t.Errorf("body missing welcome hero image, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), i18n.DefaultCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestHomeHandler_IncludesPWA(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), i18n.DefaultCatalog())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	for _, want := range []string{
		"manifest.webmanifest",
		"theme-color",
		"serviceWorker",
		"apple-mobile-web-app-capable",
		`property="og:title"`,
		`name="twitter:card"`,
		"https://cais.example.com/static/og.png",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}
