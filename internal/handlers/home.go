package handlers

import (
	"net/http"

	"github.com/matheuspuppe/cais/pkg/cais"
	"github.com/matheuspuppe/cais/pkg/cais/httpx"
)

type PageData struct {
	Nome string
}

type HomeHandler struct {
	renderer *cais.Renderer
}

func NewHomeHandler(renderer *cais.Renderer) *HomeHandler {
	return &HomeHandler{renderer: renderer}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "home", PageData{Nome: "Desenvolvedor"})
}
