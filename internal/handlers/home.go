package handlers

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type PageData struct {
	meta.Site
	Nome string
}

type HomeHandler struct {
	renderer *cais.Renderer
	site     meta.Site
	catalog  *i18n.Catalog
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewHomeHandler(renderer *cais.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config, i *inertia.Inertia) *HomeHandler {
	return &HomeHandler{renderer: renderer, site: site, catalog: catalog, cfg: cfg, inertia: i}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.inertia != nil {
		err := h.inertia.Render(w, r, "Home", inertia.Props{
			"title": "Home",
			"site":  meta.ForRequest(h.site, r),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	httpx.WritePage(w, r, h.renderer, httpx.PageConfig{
		Layout: "welcome",
		Page:   "home",
		Data: PageData{
			Site: meta.ForRequest(h.site, r),
			Nome: h.catalog.T("home.greeting"),
		},
	}, h.cfg)
}
