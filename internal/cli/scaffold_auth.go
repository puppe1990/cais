package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldAuth(dir string, data scaffoldData, dryRun bool) error {
	if _, err := os.Stat(filepath.Join(dir, "internal/handlers/auth.go")); err == nil {
		return fmt.Errorf("auth already exists — remove internal/handlers/auth.go first")
	}

	migrationPath, _, err := nextMigrationFile(dir, "auth", dryRun)
	if err != nil {
		return err
	}

	files := map[string]string{
		"internal/models/user.go":                  tplUserModel,
		"internal/handlers/auth.go":                tplAuthHandler,
		"internal/handlers/auth_test.go":           tplAuthTest,
		"internal/store/password_reset.go":         tplStorePasswordReset,
		migrationPath:                              tplMigration002Auth,
		"web/templates/pages/login.html":           tplPageLogin,
		"web/templates/pages/signup.html":          tplPageSignup,
		"web/templates/pages/forgot_password.html": tplPageForgotPassword,
		"web/templates/pages/reset_password.html":  tplPageResetPassword,
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := writeScaffoldTemplate(full, content, data, path, dryRun); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	if err := patchStoreForAuth(dir, dryRun); err != nil {
		return err
	}
	if err := patchAppForAuth(dir, dryRun); err != nil {
		return err
	}
	if err := patchRoutesForAuth(dir, dryRun); err != nil {
		return err
	}

	return nil
}

func patchStoreForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "FindUserByEmail") {
		return nil
	}

	module := readModulePath(dir)
	modelsImport := fmt.Sprintf(`"%s/internal/models"`, module)

	if !strings.Contains(content, modelsImport) {
		content = strings.Replace(content,
			`import (`,
			fmt.Sprintf(`import (
	%s`, modelsImport),
			1,
		)
	}

	if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/session") {
		content = strings.Replace(content,
			`"github.com/puppe1990/cais/pkg/cais/sqllog"`,
			`"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/cais/pkg/cais/sqllog"`,
			1,
		)
	}

	if !strings.Contains(content, `"errors"`) {
		content = strings.Replace(content,
			`import (
	"database/sql"`,
			`import (
	"database/sql"
	"errors"`,
			1,
		)
	}
	if !strings.Contains(content, `"strings"`) {
		content = strings.Replace(content,
			`"path/filepath"`,
			`"path/filepath"
	"strings"`,
			1,
		)
	}
	if !strings.Contains(content, "ErrEmailTaken") {
		content = strings.Replace(content,
			`)

type Store interface {`,
			`)

var ErrEmailTaken = errors.New("email already registered")

type Store interface {`,
			1,
		)
	}

	ifaceMarker := "\n\tClose() error"
	if !strings.Contains(content, ifaceMarker) {
		return fmt.Errorf("could not patch store interface for auth")
	}
	content = strings.Replace(content,
		ifaceMarker,
		"\n\tFindUserByEmail(email string) (models.User, error)\n\tCreateUser(email, passwordHash string) (int64, error)\n\tCreatePasswordResetToken(userID int64) (string, error)\n\tFindPasswordResetUserID(token string) (int64, bool)\n\tResetPasswordWithToken(token, passwordHash string) error\n\tSessions() session.Store"+ifaceMarker,
		1,
	)

	insert := `
func (s *SQLiteStore) FindUserByEmail(email string) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (s *SQLiteStore) CreateUser(email, passwordHash string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, passwordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, ErrEmailTaken
		}
		return 0, fmt.Errorf("create user: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) Sessions() session.Store {
	return session.NewSQLiteStore(s.db.Raw())
}
`
	content = strings.Replace(content, "\nfunc (s *SQLiteStore) Close()", insert+"\nfunc (s *SQLiteStore) Close()", 1)

	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func patchAppForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/app.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	changed := false

	if !strings.Contains(content, "LoadSession") {
		if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/session") {
			content = strings.Replace(content,
				`"github.com/puppe1990/cais/pkg/cais/middleware"`,
				`"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/cais/pkg/cais/session"`,
				1,
			)
		}
		content = strings.Replace(content,
			"r.Use(middleware.CSRF(cfg))\n",
			"r.Use(middleware.CSRF(cfg))\n\tr.Use(middleware.LoadSession(deps.Store.Sessions()))\n\tr.Use(middleware.Flash)\n",
			1,
		)
		changed = true
	} else if !strings.Contains(content, "middleware.Flash") {
		content = strings.Replace(content,
			"r.Use(middleware.LoadSession(deps.Store.Sessions()))\n",
			"r.Use(middleware.LoadSession(deps.Store.Sessions()))\n\tr.Use(middleware.Flash)\n",
			1,
		)
		changed = true
	}

	if !strings.Contains(content, "SecurityHeaders") {
		content = strings.Replace(content,
			"r.Use(middleware.Recover)\n",
			"r.Use(middleware.Recover)\n\tr.Use(middleware.SecurityHeaders(cfg))\n",
			1,
		)
		changed = true
	}

	if !changed {
		return nil
	}

	return updateScaffoldFile(path, []byte(content), "internal/app/app.go", dryRun)
}

func patchRoutesForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "NewAuthHandler") {
		return nil
	}

	if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/middleware") {
		content = strings.Replace(content,
			`"github.com/puppe1990/cais/pkg/cais"`,
			`"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"`,
			1,
		)
	}
	if !strings.Contains(content, `"net/http"`) {
		content = strings.Replace(content,
			`import (
	"github.com/puppe1990/cais/pkg/cais"`,
			`import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"`,
			1,
		)
	}

	insert := `	loginLimit := middleware.NewRateLimiter(10, cfg)
	resetLimit := middleware.NewRateLimiter(10, cfg)

	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, deps.Site, deps.Store.Sessions(), cfg, deps.Catalog)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	r.Get("/signup", auth.SignUp)
	r.Post("/signup", loginLimit.Middleware(http.HandlerFunc(auth.SignUpPost)).ServeHTTP)
	r.Get("/forgot-password", auth.ForgotPassword)
	r.Post("/forgot-password", resetLimit.Middleware(http.HandlerFunc(auth.ForgotPasswordPost)).ServeHTTP)
	r.Get("/reset-password", auth.ResetPassword)
	r.Post("/reset-password", resetLimit.Middleware(http.HandlerFunc(auth.ResetPasswordPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)

`
	content = strings.Replace(content, "func registerRoutes", insert+"func registerRoutes", 1)

	content = strings.Replace(content,
		`r.Get("/dashboard", dashboard.ServeHTTP)`,
		`r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))`,
		1,
	)

	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

const tplUserModel = `package models

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
`

const tplMigration002Auth = `CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL DEFAULT (datetime('now', '+7 days'))
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token TEXT PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const tplAuthHandler = `package handlers

import (
	"errors"
	"net/http"
	"strings"

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
`

const tplStorePasswordReset = `package store

import (
	"fmt"
	"time"

	"github.com/puppe1990/cais/pkg/cais/passwordreset"
)

func (s *SQLiteStore) CreatePasswordResetToken(userID int64) (string, error) {
	if _, err := s.db.Exec("DELETE FROM password_reset_tokens WHERE user_id = ?", userID); err != nil {
		return "", fmt.Errorf("clear reset tokens: %w", err)
	}

	token, err := passwordreset.NewToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().UTC().Add(passwordreset.DefaultTTL).Format("2006-01-02 15:04:05")
	if _, err := s.db.Exec(
		"INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt,
	); err != nil {
		return "", fmt.Errorf("insert reset token: %w", err)
	}
	return token, nil
}

func (s *SQLiteStore) FindPasswordResetUserID(token string) (int64, bool) {
	var userID int64
	err := s.db.QueryRow(
		"SELECT user_id FROM password_reset_tokens WHERE token = ? AND expires_at > datetime('now')",
		token,
	).Scan(&userID)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func (s *SQLiteStore) ResetPasswordWithToken(token, passwordHash string) error {
	userID, ok := s.FindPasswordResetUserID(token)
	if !ok {
		return fmt.Errorf("invalid or expired reset token")
	}

	tx, err := s.db.Raw().Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, userID); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM password_reset_tokens WHERE token = ?", token); err != nil {
		return fmt.Errorf("delete reset token: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset: %w", err)
	}
	return nil
}
`

const tplAuthTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestAuth_Login_redirectsWhenAuthenticated(t *testing.T) {
	s := setupTestStore(t)
	sessions := s.Sessions()
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), sessions, cais.Config{}, testCatalog())

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestAuth_LoginPost_invalidCredentials(t *testing.T) {
	s := setupTestStore(t)
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{}, testCatalog())

	form := url.Values{"email": {"nobody@example.com"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid email or password") {
		t.Errorf("body missing error: %s", rr.Body.String())
	}
}
`

const tplPageLogin = `{{"{{"}} define "title" {{"}}"}}Login{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.login_title" {{"}}"}}</h2>
  {{"{{"}} if .Error {{"}}"}}<p class="text-red-600 text-sm mb-4">{{"{{"}} .Error {{"}}"}}</p>{{"{{"}} end {{"}}"}}
  <form method="post" action="/login" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">Email</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" required />
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password">{{"{{"}} t "auth.password_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password" name="password" required />
    </div>
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
      {{"{{"}} t "auth.login_submit" {{"}}"}}
    </button>
  </form>
  <p class="text-sm text-slate-600 mt-4 text-center space-y-1">
    <span class="block">
      {{"{{"}} t "auth.signup_prompt" {{"}}"}}
      <a class="text-indigo-600 hover:text-indigo-800" href="/signup">{{"{{"}} t "auth.signup_title" {{"}}"}}</a>
    </span>
    <a class="text-indigo-600 hover:text-indigo-800" href="/forgot-password">{{"{{"}} t "auth.forgot_password" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageSignup = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "auth.signup_title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.signup_title" {{"}}"}}</h2>
  <form method="post" action="/signup" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">{{"{{"}} t "contact.email_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" value="{{"{{"}} .Email {{"}}"}}" required />
      {{"{{"}} fieldError .Errors "email" {{"}}"}}
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password">{{"{{"}} t "auth.password_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password" name="password" required />
      {{"{{"}} fieldError .Errors "password" {{"}}"}}
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password_confirmation">{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password_confirmation" name="password_confirmation" required />
      {{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}
    </div>
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
      {{"{{"}} t "auth.signup_submit" {{"}}"}}
    </button>
  </form>
  <p class="text-sm text-slate-600 mt-4 text-center">
    {{"{{"}} t "auth.login_prompt" {{"}}"}}
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageForgotPassword = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "auth.forgot_password_title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.forgot_password_title" {{"}}"}}</h2>
  <p class="text-sm text-slate-600 mb-4">{{"{{"}} t "auth.forgot_password_help" {{"}}"}}</p>
  <form method="post" action="/forgot-password" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="email">{{"{{"}} t "contact.email_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="email" id="email" name="email" value="{{"{{"}} .Email {{"}}"}}" required />
      {{"{{"}} fieldError .Errors "email" {{"}}"}}
    </div>
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
      {{"{{"}} t "auth.forgot_password_submit" {{"}}"}}
    </button>
  </form>
  <p class="text-sm text-slate-600 mt-4 text-center">
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageResetPassword = `{{"{{"}} define "title" {{"}}"}}{{"{{"}} t "auth.reset_password_title" {{"}}"}}{{"{{"}} end {{"}}"}} {{"{{"}} define "content" {{"}}"}}
<div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md mx-auto mt-10">
  <h2 class="text-2xl font-bold text-slate-800 mb-4">{{"{{"}} t "auth.reset_password_title" {{"}}"}}</h2>
  {{"{{"}} if .Error {{"}}"}}<p class="text-red-600 text-sm mb-4">{{"{{"}} .Error {{"}}"}}</p>{{"{{"}} end {{"}}"}}
  {{"{{"}} if .Token {{"}}"}}
  <form method="post" action="/reset-password" class="space-y-4">
    <input type="hidden" name="csrf_token" value="{{"{{"}} .CSRFToken {{"}}"}}" />
    <input type="hidden" name="token" value="{{"{{"}} .Token {{"}}"}}" />
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password">{{"{{"}} t "auth.password_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password" name="password" required />
      {{"{{"}} fieldError .Errors "password" {{"}}"}}
    </div>
    <div>
      <label class="block text-sm font-medium text-slate-700 mb-1" for="password_confirmation">{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}</label>
      <input class="w-full border border-slate-300 rounded-lg px-3 py-2" type="password" id="password_confirmation" name="password_confirmation" required />
      {{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}
    </div>
    <button class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition" type="submit">
      {{"{{"}} t "auth.reset_password_submit" {{"}}"}}
    </button>
  </form>
  {{"{{"}} end {{"}}"}}
  <p class="text-sm text-slate-600 mt-4 text-center">
    <a class="text-indigo-600 hover:text-indigo-800" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
{{"{{"}} end {{"}}"}}`
