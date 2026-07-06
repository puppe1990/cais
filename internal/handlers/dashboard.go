package handlers

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type DashboardData struct {
	meta.Site
	TotalContacts int64
	Env           string
}

type DashboardHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewDashboardHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *DashboardHandler {
	return &DashboardHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.CountContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.inertia != nil {
		_ = h.inertia.Render(w, r, "Dashboard", inertia.Props{
			"site":          meta.ForRequest(h.site, r),
			"totalContacts": count,
			"env":           h.cfg.Env,
		})
		return
	}
	httpx.WritePage(w, r, h.renderer, httpx.PageConfig{
		Layout: "base",
		Page:   "dashboard",
		Data: DashboardData{
			Site:          meta.ForRequest(h.site, r),
			TotalContacts: count,
			Env:           h.cfg.Env,
		},
	}, h.cfg)
}
