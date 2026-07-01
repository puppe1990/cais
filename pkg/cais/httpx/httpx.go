package httpx

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
)

// RenderPage renders a full HTML page with layout.
func RenderPage(w http.ResponseWriter, renderer *cais.Renderer, layout, page string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return renderer.Render(w, layout, page, data)
}

// RenderPartial renders an HTMX fragment.
func RenderPartial(w http.ResponseWriter, renderer *cais.Renderer, partial string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return renderer.RenderPartial(w, partial, data)
}

// SeeOther redirects with 303.
func SeeOther(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// RenderOrError writes a page or returns 500 on render error.
func RenderOrError(w http.ResponseWriter, renderer *cais.Renderer, layout, page string, data any) {
	if err := RenderPage(w, renderer, layout, page, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type RenderOptions struct {
	Layout  string
	Page    string
	Partial string
	Data    any
	Status  int
}

func RenderPageOrPartial(w http.ResponseWriter, r *http.Request, renderer *cais.Renderer, opts RenderOptions) {
	if opts.Status != 0 {
		w.WriteHeader(opts.Status)
	}
	if cais.IsHTMX(r) {
		if err := RenderPartial(w, renderer, opts.Partial, opts.Data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	RenderOrError(w, renderer, opts.Layout, opts.Page, opts.Data)
}
