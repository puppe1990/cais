package handlers

import (
	"net/http"

	"github.com/matheuspuppe/cais/pkg/cais"
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "home", PageData{Nome: "Desenvolvedor"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}