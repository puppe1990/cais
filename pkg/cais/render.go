package cais

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/puppe1990/cais/pkg/cais/forms"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func NewRendererFromDir(dir string, catalog *i18n.Catalog) (*Renderer, error) {
	return NewRenderer(os.DirFS(dir), catalog)
}

type Renderer struct {
	pages    map[string]*template.Template
	partials map[string]*template.Template
	catalog  *i18n.Catalog
}

// NewRenderer parses all templates once at boot, not per request.
// Per-request parsing adds latency and scatters template paths across handlers (harder for agents to grep).
func NewRenderer(fsys fs.FS, catalog *i18n.Catalog) (*Renderer, error) {
	if catalog == nil {
		catalog = i18n.DefaultCatalog()
	}
	r := &Renderer{
		pages:    make(map[string]*template.Template),
		partials: make(map[string]*template.Template),
		catalog:  catalog,
	}

	layouts, err := fs.Glob(fsys, "layouts/*.html")
	if err != nil {
		return nil, err
	}
	if len(layouts) == 0 {
		return nil, fmt.Errorf("no layout templates found")
	}
	sort.Strings(layouts)

	partials, err := fs.Glob(fsys, "partials/*.html")
	if err != nil {
		return nil, err
	}

	pages, err := fs.Glob(fsys, "pages/*.html")
	if err != nil {
		return nil, err
	}
	for _, pagePath := range pages {
		name := strings.TrimSuffix(filepath.Base(pagePath), ".html")
		tmpl, err := parsePage(fsys, layouts, pagePath, partials, catalog)
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", name, err)
		}
		r.pages[name] = tmpl
	}

	for _, partialPath := range partials {
		name := strings.TrimSuffix(filepath.Base(partialPath), ".html")
		tmpl, err := template.New("").Funcs(templateFuncs(catalog)).ParseFS(fsys, partialPath)
		if err != nil {
			return nil, fmt.Errorf("parse partial %s: %w", name, err)
		}
		r.partials[name] = tmpl
	}

	return r, nil
}

func (r *Renderer) Render(w io.Writer, layout, page string, data any) error {
	tmpl, ok := r.pages[page]
	if !ok {
		return fmt.Errorf("page %q not found", page)
	}
	return tmpl.ExecuteTemplate(w, layout, data)
}

func parsePage(fsys fs.FS, layoutPaths []string, pagePath string, partialPaths []string, catalog *i18n.Catalog) (*template.Template, error) {
	files := append(append([]string{}, layoutPaths...), pagePath)
	files = append(files, partialPaths...)
	return template.New("").Funcs(templateFuncs(catalog)).ParseFS(fsys, files...)
}

func templateFuncs(catalog *i18n.Catalog) template.FuncMap {
	extra := meta.TemplateFuncs()
	for k, v := range forms.Funcs() {
		extra[k] = v
	}
	return i18n.MergeFuncs(catalog, extra)
}

func (r *Renderer) RenderPartial(w io.Writer, partial string, data any) error {
	tmpl, ok := r.partials[partial]
	if !ok {
		return fmt.Errorf("partial %q not found", partial)
	}
	return tmpl.ExecuteTemplate(w, partial, data)
}
