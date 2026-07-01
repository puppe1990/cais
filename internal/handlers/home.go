package handlers

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type PageData struct {
	meta.Site
	Nome string
}

type HomeHandler struct {
	renderer *cais.Renderer
	site     meta.Site
}

func NewHomeHandler(renderer *cais.Renderer, site meta.Site) *HomeHandler {
	return &HomeHandler{renderer: renderer, site: site}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "home", PageData{
		Site: meta.WithCSRF(h.site, r),
		Nome: "Desenvolvedor",
	})
}
