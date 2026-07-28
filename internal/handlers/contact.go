package handlers

import (
	"net/http"
	"strings"

	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/cais/internal/models"
	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
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
	inertia  *inertia.Inertia
}

func NewContactHandler(renderer *cais.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config, i *inertia.Inertia) *ContactHandler {
	return &ContactHandler{renderer: renderer, store: s, site: site, catalog: catalog, cfg: cfg, inertia: i}
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	props := inertia.Props{"site": meta.ForRequest(h.site, r)}
	if msg, ok := flash.MessageFromRequest(r); ok {
		props["flash"] = inertia.Flash{msg.Kind: msg.Message}
	}
	_ = h.inertia.Render(w, r, "Contact", props)
}

func (h *ContactHandler) Post(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
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
		ve := make(inertia.ValidationErrors)
		for k, v := range errs {
			ve[k] = v
		}
		ctx := inertia.SetValidationErrors(r.Context(), ve)
		_ = h.inertia.Render(w, r.WithContext(ctx), "Contact", inertia.Props{})
		return
	}

	if _, err := h.store.InsertContact(models.Contact{Name: name, Email: email}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flash.Set(w, "success", "Message sent successfully.", h.cfg.CookieSecure())
	h.inertia.Redirect(w, r, "/contact", http.StatusSeeOther)
}
