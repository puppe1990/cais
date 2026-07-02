package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/internal/store"
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
}

type loginData struct {
	meta.Site
	Error string
}

type forgotPasswordData struct {
	meta.Site
	Email  string
	Errors validate.FieldErrors
}

type resetPasswordData struct {
	meta.Site
	Token  string
	Errors validate.FieldErrors
	Error  string
}

type signupData struct {
	meta.Site
	Email  string
	Errors validate.FieldErrors
}

func NewAuthHandler(renderer *cais.Renderer, s store.Store, site meta.Site, sessions session.Store, cfg cais.Config, catalog *i18n.Catalog) *AuthHandler {
	return &AuthHandler{renderer: renderer, store: s, site: site, sessions: sessions, cfg: cfg, catalog: catalog}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "login", loginData{Site: meta.ForRequest(h.site, r)}, h.cfg)
}

func (h *AuthHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	user, err := h.store.FindUserByEmail(email)
	if err != nil || !session.VerifyPassword(user.PasswordHash, password) {
		httpx.RenderOrError(w, h.renderer, "base", "login", loginData{
			Site:  meta.ForRequest(h.site, r),
			Error: h.catalog.T("auth.invalid_credentials"),
		}, h.cfg)
		return
	}

	if err := session.SignIn(w, h.sessions, r, user.ID, session.CookieOptionsFromConfig(h.cfg)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flash.Set(w, "notice", h.catalog.T("auth.welcome"), h.cfg.CookieSecure())
	httpx.SeeOther(w, r, "/dashboard")
}

func (h *AuthHandler) LogoutPost(w http.ResponseWriter, r *http.Request) {
	session.SignOut(w, h.sessions, r, session.CookieOptionsFromConfig(h.cfg))
	httpx.SeeOther(w, r, "/login")
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "signup", signupData{Site: meta.ForRequest(h.site, r)}, h.cfg)
}

func (h *AuthHandler) SignUpPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
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
		httpx.RenderOrError(w, h.renderer, "base", "signup", signupData{
			Site:   meta.ForRequest(h.site, r),
			Email:  email,
			Errors: errs,
		}, h.cfg)
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
			httpx.RenderOrError(w, h.renderer, "base", "signup", signupData{
				Site:  meta.ForRequest(h.site, r),
				Email: email,
				Errors: validate.FieldErrors{
					"email": h.catalog.T("auth.email_taken"),
				},
			}, h.cfg)
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
	httpx.SeeOther(w, r, "/dashboard")
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "forgot_password", forgotPasswordData{
		Site: meta.ForRequest(h.site, r),
	}, h.cfg)
}

func (h *AuthHandler) ForgotPasswordPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	var errs validate.FieldErrors
	if err := validate.Email(email); err != nil {
		errs.Add("email", h.catalog.T("contact.email_invalid"))
	}
	if errs.Any() {
		httpx.RenderOrError(w, h.renderer, "base", "forgot_password", forgotPasswordData{
			Site:   meta.ForRequest(h.site, r),
			Email:  email,
			Errors: errs,
		}, h.cfg)
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
	httpx.SeeOther(w, r, "/login")
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		httpx.RenderOrError(w, h.renderer, "base", "reset_password", resetPasswordData{
			Site:  meta.ForRequest(h.site, r),
			Error: h.catalog.T("auth.reset_invalid_token"),
		}, h.cfg)
		return
	}
	if _, ok := h.store.FindPasswordResetUserID(token); !ok {
		httpx.RenderOrError(w, h.renderer, "base", "reset_password", resetPasswordData{
			Site:  meta.ForRequest(h.site, r),
			Error: h.catalog.T("auth.reset_invalid_token"),
		}, h.cfg)
		return
	}

	httpx.RenderOrError(w, h.renderer, "base", "reset_password", resetPasswordData{
		Site:  meta.ForRequest(h.site, r),
		Token: token,
	}, h.cfg)
}

func (h *AuthHandler) ResetPasswordPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
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
		httpx.RenderOrError(w, h.renderer, "base", "reset_password", resetPasswordData{
			Site:  meta.ForRequest(h.site, r),
			Error: h.catalog.T("auth.reset_invalid_token"),
		}, h.cfg)
		return
	}
	if err := validate.MinLength(password, 8); err != nil {
		errs.Add("password", h.catalog.T("auth.password_too_short"))
	}
	if password != confirm {
		errs.Add("password_confirmation", h.catalog.T("auth.password_mismatch"))
	}
	if errs.Any() {
		httpx.RenderOrError(w, h.renderer, "base", "reset_password", resetPasswordData{
			Site:   meta.ForRequest(h.site, r),
			Token:  token,
			Errors: errs,
		}, h.cfg)
		return
	}

	hash, err := session.HashPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.ResetPasswordWithToken(token, hash); err != nil {
		httpx.RenderOrError(w, h.renderer, "base", "reset_password", resetPasswordData{
			Site:  meta.ForRequest(h.site, r),
			Error: h.catalog.T("auth.reset_invalid_token"),
		}, h.cfg)
		return
	}

	flash.Set(w, "notice", h.catalog.T("auth.reset_success"), h.cfg.CookieSecure())
	httpx.SeeOther(w, r, "/login")
}

func (h *AuthHandler) resetNotifier() passwordreset.Notifier {
	if h.resetNotify != nil {
		return h.resetNotify
	}
	return passwordreset.NotifierFromConfig(h.cfg, h.site.AppName)
}
