package handlers

import (
	"net/http"
	"strings"

	"github.com/matheuspuppe/cais/internal/models"
	"github.com/matheuspuppe/cais/internal/store"
	"github.com/matheuspuppe/cais/pkg/cais"
)

type ContactHandler struct {
	renderer *cais.Renderer
	store    store.Store
}

type contactErrorData struct {
	Message string
}

func NewContactHandler(renderer *cais.Renderer, s store.Store) *ContactHandler {
	return &ContactHandler{renderer: renderer, store: s}
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "contact", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ContactHandler) Post(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))

	if email == "" {
		h.renderContactResponse(w, r, http.StatusUnprocessableEntity, "contact_errors", contactErrorData{
			Message: "O campo email é obrigatório.",
		})
		return
	}

	_, err := h.store.InsertContact(models.Contact{Name: name, Email: email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderContactResponse(w, r, http.StatusOK, "contact_success", nil)
}

func (h *ContactHandler) renderContactResponse(w http.ResponseWriter, r *http.Request, status int, tmpl string, data any) {
	w.WriteHeader(status)
	if cais.IsHTMX(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.renderer.RenderPartial(w, tmpl, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, "base", "contact", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
