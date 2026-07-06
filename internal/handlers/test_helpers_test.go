package handlers

import (
	"os"
	"path/filepath"
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
	root := projectRoot(t)
	layout := filepath.Join(root, "web", "templates", "layouts", "base.html")
	if _, err := os.Stat(layout); err != nil {
		return cais.NewRendererStub(i18n.DefaultCatalog())
	}
	r, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), i18n.DefaultCatalog())
	if err != nil {
		t.Fatal(err)
	}
	return r
}
