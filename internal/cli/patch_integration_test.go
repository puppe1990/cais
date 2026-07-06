package cli

import (
	"strings"
	"testing"
)

func TestInsertBeforeFunctionEnd_scaffoldRoutesWithResourceGroup(t *testing.T) {
	src := `package app

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/demo/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg, nil)
	r.Get("/", home.ServeHTTP)
}
`
	insert := `
	adminItems := handlers.NewAdminItemsHandler(deps.Renderer, deps.Store, cfg)
	r.Group(middleware.AdminAuth(cfg), func(g *cais.Router) {
		g.Get("/admin/items", adminItems.Index)
		g.Get("/admin/items/{id}", cais.IntParam("id", adminItems.Show))
	})
`
	out, err := insertBeforeFunctionEnd(src, "registerRoutes", insert)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "middleware.AdminAuth(cfg)") {
		t.Fatalf("missing AdminAuth in:\n%s", out)
	}
}
