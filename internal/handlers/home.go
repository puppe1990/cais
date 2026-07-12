package handlers

import (
	"fmt"
	"net/http"

	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
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
		site := meta.ForRequest(h.site, r)
		props := inertia.Props{
			"title": h.catalog.T("home.title"),
			"site":  site,
			"labels": map[string]string{
				"heading":   h.catalog.T("home.rails_heading"),
				"subtitle":  fmt.Sprintf(h.catalog.T("home.rails_subtitle"), site.AppName),
				"stack":     h.catalog.T("home.stack"),
				"contact":   h.catalog.T("home.contact_link"),
				"login":     h.catalog.T("auth.login_submit"),
				"dashboard": h.catalog.T("dashboard.title"),
			},
		}
		if msg, ok := flash.MessageFromRequest(r); ok {
			props["flash"] = inertia.Flash{msg.Kind: msg.Message}
		}
		err := h.inertia.Render(w, r, "Home", props)
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
