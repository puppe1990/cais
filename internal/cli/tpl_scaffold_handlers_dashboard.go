// Inertia handler scaffold templates (ported from Cais demo).
package cli

const tplDashboardHandler = `package handlers

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v3"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
	"github.com/puppe1990/cais/pkg/cais/meta"

	"{{.ModulePath}}/internal/store"
)

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

	props := inertia.Props{
		"site":          meta.ForRequest(h.site, r),
		"totalContacts": count,
		"env":           h.cfg.Env,
	}
	if msg, ok := flash.MessageFromRequest(r); ok {
		props["flash"] = inertia.Flash{msg.Kind: msg.Message}
	}
	_ = h.inertia.Render(w, r, "Dashboard", props)
}
`
