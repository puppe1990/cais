package cli

import (
	"strings"
	"testing"
)

func TestInsertBeforeFunctionEnd_insertsRoutes(t *testing.T) {
	src := `package app

func registerRoutes(r *cais.Router) {
	r.Get("/", home)
}

func other() {}
`
	out, err := insertBeforeFunctionEnd(src, "registerRoutes", "\tr.Get(\"/new\", h)\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `r.Get("/new", h)`) {
		t.Error("missing insert")
	}
	if strings.Index(out, `r.Get("/new"`) > strings.Index(out, "func other") {
		t.Error("insert should be inside registerRoutes before other func")
	}
}

func TestInsertBeforeFunctionEnd_nestedBraces(t *testing.T) {
	src := `package app

func registerRoutes(r *cais.Router) {
	r.Group(func(g *cais.Router) {
		g.Get("/", home)
	})
}

func other() {}
`
	out, err := insertBeforeFunctionEnd(src, "registerRoutes", "\tr.Get(\"/tail\", tail)\n")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(out, `r.Get("/tail"`) > strings.Index(out, "func other") {
		t.Error("insert should be inside registerRoutes")
	}
	if strings.Index(out, `r.Get("/tail"`) < strings.Index(out, `g.Get("/", home)`) {
		t.Error("insert should be after nested group")
	}
}

func TestInsertBeforeFunctionEnd_notFound(t *testing.T) {
	_, err := insertBeforeFunctionEnd("package app\n", "registerRoutes", "x")
	if err == nil {
		t.Fatal("expected error for missing function")
	}
}
