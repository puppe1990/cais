// Inertia handler scaffold templates (ported from Cais demo).
package cli

const tplHomeHandler = `package handlers

import (
	"fmt"
	"net/http"

	inertia "github.com/romsar/gonertia/v3"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

type PageData struct {
	meta.Site
	Nome string
}

type HomeHandler struct {
	renderer *cais.Renderer
	site     meta.Site
	catalog  *i18n.Catalog
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewHomeHandler(renderer *cais.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config, i *inertia.Inertia) *HomeHandler {
	return &HomeHandler{renderer: renderer, site: site, catalog: catalog, cfg: cfg, inertia: i}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.inertia != nil {
		site := meta.ForRequest(h.site, r)
		props := inertia.Props{
			"title": h.catalog.T("home.title"),
			"site":  site,
			"labels": map[string]string{
				"heading":   h.catalog.T("home.rails_heading"),
				"subtitle":  fmt.Sprintf(h.catalog.T("home.rails_subtitle"), site.AppName),
				"stack":     h.catalog.T("home.stack"),
				"contact":   h.catalog.T("home.contact_link"),
				"login":     h.catalog.T("auth.login_submit"),
				"dashboard": h.catalog.T("dashboard.title"),
			},
		}
		if msg, ok := flash.MessageFromRequest(r); ok {
			props["flash"] = inertia.Flash{msg.Kind: msg.Message}
		}
		err := h.inertia.Render(w, r, "Home", props)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	httpx.RenderOrError(w, h.renderer, "welcome", "home", PageData{
		Site: meta.ForRequest(h.site, r),
	}, h.cfg)
}
`

const tplContactHandler = `package handlers

import (
	"net/http"
	"strings"

	inertia "github.com/romsar/gonertia/v3"
	"{{.ModulePath}}/internal/models"
	"{{.ModulePath}}/internal/store"
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

type contactErrorData struct {
	Message string
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.inertia != nil {
		props := inertia.Props{"site": meta.ForRequest(h.site, r)}
		if msg, ok := flash.MessageFromRequest(r); ok {
			props["flash"] = inertia.Flash{msg.Kind: msg.Message}
		}
		_ = h.inertia.Render(w, r, "Contact", props)
		return
	}
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
		if h.inertia != nil {
			ve := make(inertia.ValidationErrors)
			for k, v := range errs {
				ve[k] = v
			}
			ctx := inertia.SetValidationErrors(r.Context(), ve)
			// render same component so props.errors populated by gonertia
			_ = h.inertia.Render(w, r.WithContext(ctx), "Contact", inertia.Props{})
			return
		}
		h.renderContactResponse(w, r, http.StatusUnprocessableEntity, "contact_errors", contactErrorData{Message: errs.First()})
		return
	}

	_, err := h.store.InsertContact(models.Contact{Name: name, Email: email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.inertia != nil {
		flash.Set(w, "success", "Message sent successfully.", h.cfg.CookieSecure())
		h.inertia.Redirect(w, r, "/contact", http.StatusSeeOther)
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
`

const tplDashboardHandler = `package handlers

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v3"
	"{{.ModulePath}}/internal/store"
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
	httpx.RenderOrError(w, h.renderer, "base", "dashboard", DashboardData{
		Site:          meta.ForRequest(h.site, r),
		TotalContacts: count,
		Env:           h.cfg.Env,
	}, h.cfg)
}
`

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

// keep data types for fallback render paths (old templates)
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

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		if h.inertia != nil {
			h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if h.inertia != nil {
		_ = h.inertia.Render(w, r, "Login", inertia.Props{
			"site": meta.ForRequest(h.site, r),
		})
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "login", nil, h.cfg)
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
		if h.inertia != nil {
			ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
				"email": h.catalog.T("auth.invalid_credentials"),
			})
			_ = h.inertia.Render(w, r.WithContext(ctx), "Login", inertia.Props{})
			return
		}
		httpx.RenderOrError(w, h.renderer, "base", "login", nil, h.cfg)
		return
	}

	if err := session.SignIn(w, h.sessions, r, user.ID, session.CookieOptionsFromConfig(h.cfg)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.inertia != nil {
		ctx := inertia.SetFlash(r.Context(), inertia.Flash{"notice": h.catalog.T("auth.welcome")})
		h.inertia.Redirect(w, r.WithContext(ctx), "/dashboard", http.StatusSeeOther)
		return
	}
	flash.Set(w, "notice", h.catalog.T("auth.welcome"), h.cfg.CookieSecure())
	httpx.SeeOther(w, r, "/dashboard")
}

func (h *AuthHandler) LogoutPost(w http.ResponseWriter, r *http.Request) {
	session.SignOut(w, h.sessions, r, session.CookieOptionsFromConfig(h.cfg))
	if h.inertia != nil {
		h.inertia.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	httpx.SeeOther(w, r, "/login")
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		if h.inertia != nil {
			h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if h.inertia != nil {
		_ = h.inertia.Render(w, r, "Signup", inertia.Props{"site": meta.ForRequest(h.site, r)})
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
		if h.inertia != nil {
			ve := make(inertia.ValidationErrors)
			for k, v := range errs {
				ve[k] = v
			}
			ctx := inertia.SetValidationErrors(r.Context(), ve)
			_ = h.inertia.Render(w, r.WithContext(ctx), "Signup", inertia.Props{})
			return
		}
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
			if h.inertia != nil {
				ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
					"email": h.catalog.T("auth.email_taken"),
				})
				_ = h.inertia.Render(w, r.WithContext(ctx), "Signup", inertia.Props{})
				return
			}
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
	if h.inertia != nil {
		ctx := inertia.SetFlash(r.Context(), inertia.Flash{"notice": h.catalog.T("auth.welcome")})
		h.inertia.Redirect(w, r.WithContext(ctx), "/dashboard", http.StatusSeeOther)
		return
	}
	flash.Set(w, "notice", h.catalog.T("auth.welcome"), h.cfg.CookieSecure())
	httpx.SeeOther(w, r, "/dashboard")
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		if h.inertia != nil {
			h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if h.inertia != nil {
		_ = h.inertia.Render(w, r, "ForgotPassword", inertia.Props{"site": meta.ForRequest(h.site, r)})
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
		if h.inertia != nil {
			ve := make(inertia.ValidationErrors)
			for k, v := range errs {
				ve[k] = v
			}
			ctx := inertia.SetValidationErrors(r.Context(), ve)
			_ = h.inertia.Render(w, r.WithContext(ctx), "ForgotPassword", inertia.Props{})
			return
		}
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
	if h.inertia != nil {
		h.inertia.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	httpx.SeeOther(w, r, "/login")
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); ok {
		if h.inertia != nil {
			h.inertia.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if h.inertia != nil {
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
		return
	}

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
		if h.inertia != nil {
			ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
				"token": h.catalog.T("auth.reset_invalid_token"),
			})
			_ = h.inertia.Render(w, r.WithContext(ctx), "ResetPassword", inertia.Props{
				"site":  meta.ForRequest(h.site, r),
				"token": token,
			})
			return
		}
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
		if h.inertia != nil {
			ve := make(inertia.ValidationErrors)
			for k, v := range errs {
				ve[k] = v
			}
			ctx := inertia.SetValidationErrors(r.Context(), ve)
			_ = h.inertia.Render(w, r.WithContext(ctx), "ResetPassword", inertia.Props{"token": token})
			return
		}
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

	if h.inertia != nil {
		ctx := inertia.SetFlash(r.Context(), inertia.Flash{"notice": h.catalog.T("auth.reset_success")})
		h.inertia.Redirect(w, r.WithContext(ctx), "/login", http.StatusSeeOther)
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
`

const tplHelpersTest = `package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func testSite() meta.Site {
	return meta.Site{AppName: "{{.AppName}}", AppURL: "https://cais.example.com"}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func setupTestRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	root := projectRoot(t)
	layout := filepath.Join(root, "web", "templates", "layouts", "base.html")
	if _, err := os.Stat(layout); err != nil {
		return cais.NewRendererStub(i18n.DefaultCatalog())
	}
	r, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), i18n.DefaultCatalog())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}`

const tplInertiaTest = `package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	inertia "github.com/romsar/gonertia/v3"
)

const testInertiaRoot = ` + "`" + `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8" />{{"{{ .inertiaHead }}"}}</head>
<body>{{"{{ .inertia }}"}}</body>
</html>` + "`" + `

func setupTestInertia(t *testing.T) *inertia.Inertia {
	t.Helper()
	i, err := inertia.New(testInertiaRoot)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func inertiaRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("X-Inertia", "true")
	return req
}

func parseInertiaJSON(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("not json: %v body=%s", err, rr.Body.String())
	}
	return payload
}

func assertInertiaComponent(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	if payload["component"] != want {
		t.Errorf("component = %v, want %s", payload["component"], want)
	}
}

func assertInertiaErrors(t *testing.T, rr *httptest.ResponseRecorder, keys ...string) {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	errors, ok := props["errors"].(map[string]any)
	if !ok || len(errors) == 0 {
		t.Fatalf("missing errors in props: %v", props)
	}
	for _, k := range keys {
		if _, ok := errors[k]; !ok {
			t.Errorf("errors missing key %q: %v", k, errors)
		}
	}
}

func assertInertiaProp(t *testing.T, rr *httptest.ResponseRecorder, key string) any {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	v, ok := props[key]
	if !ok {
		t.Fatalf("props missing %q: %v", key, props)
	}
	return v
}`

const tplHomeTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
)

func newHomeHandler(t *testing.T) *HomeHandler {
	t.Helper()
	return NewHomeHandler(setupTestRenderer(t), testSite(), i18n.DefaultCatalog(), cais.Config{}, setupTestInertia(t))
}

func TestHomeHandler_Returns200(t *testing.T) {
	h := newHomeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_InertiaComponent(t *testing.T) {
	h := newHomeHandler(t)

	req := inertiaRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assertInertiaComponent(t, rr, "Home")
}

func TestHomeHandler_InertiaShell(t *testing.T) {
	h := newHomeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, ` + "`" + `id="app"` + "`" + `) && !strings.Contains(body, "data-page") {
		t.Errorf("body missing Inertia shell markers, got: %s", body)
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := newHomeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}`

const tplContactTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
	"github.com/puppe1990/cais/pkg/cais/i18n"
)

func newContactHandler(t *testing.T) (*ContactHandler, store.Store) {
	t.Helper()
	s := setupTestStore(t)
	h := NewContactHandler(setupTestRenderer(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{}, setupTestInertia(t))
	return h, s
}

func TestContactHandler_Get_InertiaComponent(t *testing.T) {
	h, _ := newContactHandler(t)

	req := inertiaRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	assertInertiaComponent(t, rr, "Contact")
}

func TestContactHandler_Post_MalformedEmail_InertiaErrors(t *testing.T) {
	h, _ := newContactHandler(t)

	req := inertiaRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=not-an-email"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	assertInertiaComponent(t, rr, "Contact")
	assertInertiaErrors(t, rr, "email")
}

func TestContactHandler_Post_MissingName_InertiaErrors(t *testing.T) {
	h, _ := newContactHandler(t)

	req := inertiaRequest(http.MethodPost, "/contact", strings.NewReader("name=&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	assertInertiaErrors(t, rr, "name")
}

func TestContactHandler_Post_InvalidEmail_InertiaErrors(t *testing.T) {
	h, _ := newContactHandler(t)

	req := inertiaRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	assertInertiaErrors(t, rr, "email")
}

func TestContactHandler_Get_InertiaFlash(t *testing.T) {
	h, _ := newContactHandler(t)

	req := inertiaRequest(http.MethodGet, "/contact", nil)
	req = flash.WithMessage(req, flash.Message{Kind: "success", Message: "Message sent successfully."})
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	payload := parseInertiaJSON(t, rr)
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	flashProp, ok := props["flash"].(map[string]any)
	if !ok || flashProp["success"] != "Message sent successfully." {
		t.Errorf("props.flash missing success: %v", props)
	}
}

func TestContactHandler_Post_Valid_Redirects(t *testing.T) {
	h, s := newContactHandler(t)

	req := inertiaRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if rr.Header().Get("Location") != "/contact" {
		t.Errorf("Location = %q, want /contact", rr.Header().Get("Location"))
	}
	count, err := s.CountContacts()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("contact count = %d, want 1", count)
	}
}`

const tplAuthTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func newAuthHandler(t *testing.T) (*AuthHandler, store.Store) {
	t.Helper()
	s := setupTestStore(t)
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog(), setupTestInertia(t))
	return h, s
}

func TestAuth_Login_redirectsWhenAuthenticated(t *testing.T) {
	h, s := newAuthHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	_ = s
}

func TestAuth_LoginPost_invalidCredentials(t *testing.T) {
	h, _ := newAuthHandler(t)

	form := url.Values{"email": {"nobody@example.com"}, "password": {"wrong"}}
	req := inertiaRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	assertInertiaComponent(t, rr, "Login")
	assertInertiaErrors(t, rr, "email")
}

func TestAuth_LoginPost_validCredentials_redirects(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog(), setupTestInertia(t))

	form := url.Values{"email": {"demo@example.com"}, "password": {"password"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", rr.Header().Get("Location"))
	}
}`

const tplAuthSignupTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
)

func newAuthHandlerForSignup(t *testing.T) (*AuthHandler, store.Store) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog(), setupTestInertia(t))
	return h, s
}

func TestAuth_SignUpPost_createsUserAndRedirects(t *testing.T) {
	h, s := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", rr.Header().Get("Location"))
	}

	user, err := s.FindUserByEmail("signup@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Fatal("user id = 0")
	}
}

func TestAuth_SignUpPost_duplicateEmail_returnsError(t *testing.T) {
	h, _ := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("first signup status = %d, want 303", rr.Code)
	}

	req2 := inertiaRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr2 := httptest.NewRecorder()
	h.SignUpPost(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("duplicate signup status = %d, want 200", rr2.Code)
	}
	assertInertiaComponent(t, rr2, "Signup")
	assertInertiaErrors(t, rr2, "email")
}

func TestAuth_SignUp_InertiaComponent(t *testing.T) {
	h, _ := newAuthHandlerForSignup(t)

	req := inertiaRequest(http.MethodGet, "/signup", nil)
	rr := httptest.NewRecorder()
	h.SignUp(rr, req)

	assertInertiaComponent(t, rr, "Signup")
}`

const tplAuthResetTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/passwordreset"
	"github.com/puppe1990/cais/pkg/cais/session"
)

type captureNotifier struct {
	emails []string
	tokens []string
}

func (c *captureNotifier) NotifyReset(email, token string) error {
	c.emails = append(c.emails, email)
	c.tokens = append(c.tokens, token)
	return nil
}

func newAuthHandlerForReset(t *testing.T, s store.Store, notify passwordreset.Notifier) *AuthHandler {
	t.Helper()
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{AppURL: "http://localhost:8080"}, i18n.DefaultCatalog(), setupTestInertia(t))
	h.resetNotify = notify
	return h
}

func TestAuth_ForgotPasswordPost_unknownEmail_showsSameMessage(t *testing.T) {
	s := setupTestStore(t)
	notify := &captureNotifier{}
	h := newAuthHandlerForReset(t, s, notify)

	form := url.Values{"email": {"missing@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ForgotPasswordPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if len(notify.emails) != 0 {
		t.Fatal("should not notify for unknown email")
	}
}

func TestAuth_ForgotPasswordPost_knownEmail_notifiesAndRedirects(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	notify := &captureNotifier{}
	h := newAuthHandlerForReset(t, s, notify)

	form := url.Values{"email": {"demo@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ForgotPasswordPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if len(notify.emails) != 1 || notify.emails[0] != "demo@example.com" {
		t.Fatalf("notify emails = %v", notify.emails)
	}
	if len(notify.tokens) != 1 || notify.tokens[0] == "" {
		t.Fatalf("notify tokens = %v", notify.tokens)
	}
}

func TestAuth_ForgotPassword_InertiaComponent(t *testing.T) {
	s := setupTestStore(t)
	h := newAuthHandlerForReset(t, s, &captureNotifier{})

	req := inertiaRequest(http.MethodGet, "/forgot-password", nil)
	rr := httptest.NewRecorder()
	h.ForgotPassword(rr, req)

	assertInertiaComponent(t, rr, "ForgotPassword")
}

func TestAuth_ResetPasswordPost_validToken_updatesPassword(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	token, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	h := newAuthHandlerForReset(t, s, &captureNotifier{})
	form := url.Values{
		"token":                 {token},
		"password":              {"new-password-123"},
		"password_confirmation": {"new-password-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ResetPasswordPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rr.Header().Get("Location"))
	}

	updated, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(updated.PasswordHash, "new-password-123") {
		t.Fatal("password was not updated")
	}
}

func TestAuth_ResetPasswordPost_invalidToken_rendersError(t *testing.T) {
	s := setupTestStore(t)
	h := newAuthHandlerForReset(t, s, &captureNotifier{})

	form := url.Values{
		"token":                 {"bad-token"},
		"password":              {"new-password-123"},
		"password_confirmation": {"new-password-123"},
	}
	req := inertiaRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ResetPasswordPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	assertInertiaComponent(t, rr, "ResetPassword")
	assertInertiaErrors(t, rr, "token")
}`

const tplDashboardTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestDashboardHandler_InertiaComponent(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{}, setupTestInertia(t))

	req := inertiaRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	assertInertiaComponent(t, rr, "Dashboard")
}
`
