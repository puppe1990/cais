// Inertia handler scaffold templates (ported from Cais demo).
package cli

const tplAuthHandler = `package handlers

import (
	"errors"
	"net/http"
	"strings"

	inertia "github.com/romsar/gonertia/v3"
	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/passwordreset"
	"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/cais/pkg/cais/validate"
)

type AuthHandler struct {
	renderer    *cais.Renderer
	store       store.Store
	site        meta.Site
	sessions    session.Store
	cfg         cais.Config
	catalog     *i18n.Catalog
	resetNotify passwordreset.Notifier
	inertia     *inertia.Inertia
}

func NewAuthHandler(renderer *cais.Renderer, s store.Store, site meta.Site, sessions session.Store, cfg cais.Config, catalog *i18n.Catalog, i *inertia.Inertia) *AuthHandler {
	return &AuthHandler{renderer: renderer, store: s, site: site, sessions: sessions, cfg: cfg, catalog: catalog, inertia: i}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	props := inertia.Props{"site": meta.ForRequest(h.site, r)}
	if msg, ok := flash.MessageFromRequest(r); ok {
		props["flash"] = inertia.Flash{msg.Kind: msg.Message}
	}
	_ = h.inertia.Render(w, r, "Login", props)
}

func (h *AuthHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	user, err := h.store.FindUserByEmail(email)
	if err != nil || !session.VerifyPassword(user.PasswordHash, password) {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"email": h.catalog.T("auth.invalid_credentials"),
		})
		_ = h.inertia.Render(w, r.WithContext(ctx), "Login", inertia.Props{})
		return
	}

	if err := session.SignIn(w, h.sessions, r, user.ID, session.CookieOptionsFromConfig(h.cfg)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// cais flash cookie (gonertia SetFlash needs FlashDataProvider; scaffold uses cookies — #140).
	flash.Set(w, "notice", h.catalog.T("auth.welcome"), h.cfg.CookieSecure())
	h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) LogoutPost(w http.ResponseWriter, r *http.Request) {
	session.SignOut(w, h.sessions, r, session.CookieOptionsFromConfig(h.cfg))
	h.inertia.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	_ = h.inertia.Render(w, r, "Signup", inertia.Props{"site": meta.ForRequest(h.site, r)})
}

func (h *AuthHandler) SignUpPost(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirmation")

	var errs validate.FieldErrors
	if err := validate.Email(email); err != nil {
		errs.Add("email", h.catalog.T("contact.email_invalid"))
	}
	if err := validate.MinLength(password, 8); err != nil {
		errs.Add("password", h.catalog.T("auth.password_too_short"))
	}
	if password != confirm {
		errs.Add("password_confirmation", h.catalog.T("auth.password_mismatch"))
	}
	if errs.Any() {
		ve := make(inertia.ValidationErrors)
		for k, v := range errs {
			ve[k] = v
		}
		ctx := inertia.SetValidationErrors(r.Context(), ve)
		_ = h.inertia.Render(w, r.WithContext(ctx), "Signup", inertia.Props{})
		return
	}

	hash, err := session.HashPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userID, err := h.store.CreateUser(email, hash)
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
				"email": h.catalog.T("auth.email_taken"),
			})
			_ = h.inertia.Render(w, r.WithContext(ctx), "Signup", inertia.Props{})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := session.SignIn(w, h.sessions, r, userID, session.CookieOptionsFromConfig(h.cfg)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flash.Set(w, "notice", h.catalog.T("auth.welcome"), h.cfg.CookieSecure())
	h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	_ = h.inertia.Render(w, r, "ForgotPassword", inertia.Props{"site": meta.ForRequest(h.site, r)})
}

func (h *AuthHandler) ForgotPasswordPost(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	var errs validate.FieldErrors
	if err := validate.Email(email); err != nil {
		errs.Add("email", h.catalog.T("contact.email_invalid"))
	}
	if errs.Any() {
		ve := make(inertia.ValidationErrors)
		for k, v := range errs {
			ve[k] = v
		}
		ctx := inertia.SetValidationErrors(r.Context(), ve)
		_ = h.inertia.Render(w, r.WithContext(ctx), "ForgotPassword", inertia.Props{})
		return
	}

	if user, err := h.store.FindUserByEmail(email); err == nil {
		token, err := h.store.CreatePasswordResetToken(user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = h.resetNotifier().NotifyReset(user.Email, token)
	}

	flash.Set(w, "notice", h.catalog.T("auth.reset_email_sent"), h.cfg.CookieSecure())
	h.inertia.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	props := inertia.Props{"site": meta.ForRequest(h.site, r), "token": token}
	if token == "" {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"token": h.catalog.T("auth.reset_invalid_token"),
		})
		_ = h.inertia.Render(w, r.WithContext(ctx), "ResetPassword", props)
		return
	}
	if _, ok := h.store.FindPasswordResetUserID(token); !ok {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"token": h.catalog.T("auth.reset_invalid_token"),
		})
		_ = h.inertia.Render(w, r.WithContext(ctx), "ResetPassword", props)
		return
	}
	_ = h.inertia.Render(w, r, "ResetPassword", props)
}

func (h *AuthHandler) ResetPasswordPost(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := strings.TrimSpace(r.FormValue("token"))
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirmation")

	var errs validate.FieldErrors
	if token == "" {
		errs.Add("token", h.catalog.T("auth.reset_invalid_token"))
	} else if _, ok := h.store.FindPasswordResetUserID(token); !ok {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"token": h.catalog.T("auth.reset_invalid_token"),
		})
		_ = h.inertia.Render(w, r.WithContext(ctx), "ResetPassword", inertia.Props{
			"site":  meta.ForRequest(h.site, r),
			"token": token,
		})
		return
	}
	if err := validate.MinLength(password, 8); err != nil {
		errs.Add("password", h.catalog.T("auth.password_too_short"))
	}
	if password != confirm {
		errs.Add("password_confirmation", h.catalog.T("auth.password_mismatch"))
	}
	if errs.Any() {
		ve := make(inertia.ValidationErrors)
		for k, v := range errs {
			ve[k] = v
		}
		ctx := inertia.SetValidationErrors(r.Context(), ve)
		_ = h.inertia.Render(w, r.WithContext(ctx), "ResetPassword", inertia.Props{"token": token})
		return
	}

	hash, err := session.HashPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.ResetPasswordWithToken(token, hash); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"token": h.catalog.T("auth.reset_invalid_token"),
		})
		_ = h.inertia.Render(w, r.WithContext(ctx), "ResetPassword", inertia.Props{
			"site":  meta.ForRequest(h.site, r),
			"token": token,
		})
		return
	}

	flash.Set(w, "notice", h.catalog.T("auth.reset_success"), h.cfg.CookieSecure())
	h.inertia.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) resetNotifier() passwordreset.Notifier {
	if h.resetNotify != nil {
		return h.resetNotify
	}
	return passwordreset.NotifierFromConfig(h.cfg, h.site.AppName)
}
`
