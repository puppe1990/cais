package handlers

import (
	"net/http"
	"strings"

	"github.com/puppe1990/cais/internal/models"
	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/validate"
)

type ContactHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	catalog  *i18n.Catalog
	cfg      cais.Config
}

type contactErrorData struct {
	Message string
}

func NewContactHandler(renderer *cais.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *ContactHandler {
	return &ContactHandler{renderer: renderer, store: s, site: site, catalog: catalog, cfg: cfg}
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	httpx.RenderOrError(w, h.renderer, "base", "contact", meta.ForRequest(h.site, r), h.cfg)
}

func (h *ContactHandler) Post(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))

	var errs validate.FieldErrors
	if name == "" {
		errs.Add("name", h.catalog.T("contact.name_required"))
	}
	if err := validate.Email(email); err != nil {
		msg := h.catalog.T("contact.email_required")
		if email != "" {
			msg = h.catalog.T("contact.email_invalid")
		}
		errs.Add("email", msg)
	}
	if errs.Any() {
		h.renderContactResponse(w, r, http.StatusUnprocessableEntity, "contact_errors", contactErrorData{Message: errs.First()})
		return
	}

	_, err := h.store.InsertContact(models.Contact{Name: name, Email: email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderContactResponse(w, r, http.StatusOK, "contact_success", nil)
}

func (h *ContactHandler) renderContactResponse(w http.ResponseWriter, r *http.Request, status int, partial string, data any) {
	httpx.RenderPageOrPartial(w, r, h.renderer, httpx.RenderOptions{
		Layout:  "base",
		Page:    "contact",
		Partial: partial,
		Data:    data,
		Status:  status,
	}, h.cfg)
}
