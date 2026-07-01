package cais

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func NewRendererFromDir(dir string) (*Renderer, error) {
	return NewRenderer(os.DirFS(dir))
}

type Renderer struct {
	pages    map[string]*template.Template
	partials map[string]*template.Template
}

func NewRenderer(fsys fs.FS) (*Renderer, error) {
	r := &Renderer{
		pages:    make(map[string]*template.Template),
		partials: make(map[string]*template.Template),
	}

	layouts, err := fs.Glob(fsys, "layouts/*.html")
	if err != nil {
		return nil, err
	}
	if len(layouts) == 0 {
		return nil, fmt.Errorf("no layout templates found")
	}
	layoutPath := layouts[0]

	pages, err := fs.Glob(fsys, "pages/*.html")
	if err != nil {
		return nil, err
	}
	for _, pagePath := range pages {
		name := strings.TrimSuffix(filepath.Base(pagePath), ".html")
		tmpl, err := template.ParseFS(fsys, layoutPath, pagePath)
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", name, err)
		}
		r.pages[name] = tmpl
	}

	partials, err := fs.Glob(fsys, "partials/*.html")
	if err != nil {
		return nil, err
	}
	for _, partialPath := range partials {
		name := strings.TrimSuffix(filepath.Base(partialPath), ".html")
		tmpl, err := template.ParseFS(fsys, partialPath)
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

func (r *Renderer) RenderPartial(w io.Writer, partial string, data any) error {
	tmpl, ok := r.partials[partial]
	if !ok {
		return fmt.Errorf("partial %q not found", partial)
	}
	return tmpl.ExecuteTemplate(w, partial, data)
}
