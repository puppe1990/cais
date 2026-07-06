package patch

import (
	"strings"
	"testing"
)

const sampleRoutes = `package app

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg, nil)
	r.Get("/", home.ServeHTTP)
}
`

func TestInsertBeforeFuncEnd_addsStatements(t *testing.T) {
	insert := `
	about := handlers.NewAboutHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)
	r.Get("/about", about.ServeHTTP)
`
	out, err := InsertBeforeFuncEnd([]byte(sampleRoutes), "registerRoutes", insert)
	if err != nil {
		t.Fatal(err)
	}
	body := string(out)
	if !strings.Contains(body, `r.Get("/about", about.ServeHTTP)`) {
		t.Errorf("missing inserted route:\n%s", body)
	}
	if !strings.Contains(body, `r.Get("/", home.ServeHTTP)`) {
		t.Error("original route should remain")
	}
}

func TestInsertBeforeFuncEnd_preservesIntParamInGroup(t *testing.T) {
	insert := `
	adminItems := handlers.NewAdminItemsHandler(deps.Renderer, deps.Store, cfg)
	r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
		g.Get("/admin/items/{id}", cais.IntParam("id", adminItems.Show))
		g.Get("/admin/items/{id}/edit", cais.IntParam("id", adminItems.Edit))
	})
`
	out, err := InsertBeforeFuncEnd([]byte(sampleRoutes), "registerRoutes", insert)
	if err != nil {
		t.Fatal(err)
	}
	body := string(out)
	for _, want := range []string{
		`cais.IntParam("id", adminItems.Show)`,
		`cais.IntParam("id", adminItems.Edit)`,
		`middleware.RequireAuth("/login")`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in:\n%s", want, body)
		}
	}
	if strings.Contains(body, "cais.\n") || strings.Contains(body, "middleware.\n") {
		t.Fatalf("qualified calls should not be split across lines:\n%s", body)
	}
}
