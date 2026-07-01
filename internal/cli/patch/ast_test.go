package patch

import (
	"strings"
	"testing"
)

const sampleRoutes = `package app

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)
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
