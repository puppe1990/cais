package testutil

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

// ProjectRoot walks up from cwd to find go.mod.
func ProjectRoot(t *testing.T) string {
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

// NewRenderer loads templates from web/templates relative to module root.
func NewRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	root := ProjectRoot(t)
	r, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

type RequestOption func(*http.Request)

// PathValue sets a Go 1.22+ path parameter on the request.
func PathValue(key, value string) RequestOption {
	return func(r *http.Request) {
		r.SetPathValue(key, value)
	}
}

// NewRequest builds an httptest request with optional path values.
func NewRequest(method, target string, opts ...RequestOption) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	for _, opt := range opts {
		opt(req)
	}
	return req
}
