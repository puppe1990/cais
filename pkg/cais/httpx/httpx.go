package httpx

import (
	"log"
	"net/http"
	"strings"

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

func writeRenderError(w http.ResponseWriter, err error, cfg cais.Config) {
	if cfg.SanitizeErrors() {
		log.Printf("render error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// RenderOrError writes a page or returns 500 on render error.
func RenderOrError(w http.ResponseWriter, renderer *cais.Renderer, layout, page string, data any, cfg cais.Config) {
	if err := RenderPage(w, renderer, layout, page, data); err != nil {
		writeRenderError(w, err, cfg)
	}
}

type RenderOptions struct {
	Layout  string
	Page    string
	Partial string
	Data    any
	Status  int
}

func RenderPageOrPartial(w http.ResponseWriter, r *http.Request, renderer *cais.Renderer, opts RenderOptions, cfg cais.Config) {
	if opts.Status != 0 {
		w.WriteHeader(opts.Status)
	}
	if cais.IsHTMX(r) {
		if err := RenderPartial(w, renderer, opts.Partial, opts.Data); err != nil {
			writeRenderError(w, err, cfg)
		}
		return
	}
	RenderOrError(w, renderer, opts.Layout, opts.Page, opts.Data, cfg)
}

// NotModified returns true and sends 304 Not Modified (with ETag) when the
// request's If-None-Match matches the given etag. This is useful for list pages
// and other cacheable responses.
//
// Typical usage:
//
//	etag := `"` + cache.Hash(myListVersion) + `"`
//	if httpx.NotModified(w, r, etag) {
//	    return
//	}
//	httpx.SetETag(w, etag)
//	... render ...
func NotModified(w http.ResponseWriter, r *http.Request, etag string) bool {
	if etag == "" {
		return false
	}
	if match := r.Header.Get("If-None-Match"); match != "" {
		if stripQuotes(match) == stripQuotes(etag) {
			w.Header().Set("ETag", quoteETag(etag))
			w.WriteHeader(http.StatusNotModified)
			return true
		}
	}
	return false
}

// SetETag sets the ETag header (adding quotes if necessary).
func SetETag(w http.ResponseWriter, etag string) {
	if etag != "" {
		w.Header().Set("ETag", quoteETag(etag))
	}
}

func quoteETag(etag string) string {
	if strings.HasPrefix(etag, `"`) || strings.HasPrefix(etag, `W/"`) {
		return etag
	}
	return `"` + etag + `"`
}

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, `W/`)
	return strings.Trim(s, `"`)
}
