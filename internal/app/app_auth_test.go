package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/csrf"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
)

func TestApp_LoginPost_requiresCSRF(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=demo@example.com&password=password"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF status = %d, want 403", postRR.Code)
	}
}

func TestApp_LoginPost_withCSRF_redirects(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	var token string
	for _, c := range getRR.Result().Cookies() {
		if c.Name == csrf.CookieName {
			token = c.Value
		}
	}
	if token == "" {
		t.Fatal("missing csrf cookie after GET /login")
	}

	form := url.Values{}
	form.Set("email", "demo@example.com")
	form.Set("password", "password")
	form.Set("csrf_token", token)

	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusSeeOther {
		t.Errorf("POST with CSRF status = %d, want 303, body: %s", postRR.Code, postRR.Body.String())
	}
}

func TestApp_Dashboard_requiresAuth(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rr.Header().Get("Location"))
	}
}

func TestApp_AuthFlow_loginDashboardLogout(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	getLogin := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRR := httptest.NewRecorder()
	h.ServeHTTP(loginRR, getLogin)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d", loginRR.Code)
	}
	csrfToken := csrfTokenFromResponse(t, loginRR.Result())

	form := url.Values{}
	form.Set("email", "demo@example.com")
	form.Set("password", "password")
	form.Set("csrf_token", csrfToken)

	postLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postLogin.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	loginPostRR := httptest.NewRecorder()
	h.ServeHTTP(loginPostRR, postLogin)
	if loginPostRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /login status = %d, want 303", loginPostRR.Code)
	}
	sessionCookie := sessionCookieFromResponse(t, loginPostRR.Result())
	if sessionCookie == nil {
		t.Fatal("missing session cookie after login")
	}

	dashReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	for _, c := range loginPostRR.Result().Cookies() {
		dashReq.AddCookie(c)
	}
	dashRR := httptest.NewRecorder()
	h.ServeHTTP(dashRR, dashReq)
	if dashRR.Code != http.StatusOK {
		t.Fatalf("GET /dashboard status = %d, want 200", dashRR.Code)
	}
	// dashboard now Inertia; check markers or props (flash delivered via gonertia)
	dbody := dashRR.Body.String()
	if !strings.Contains(dbody, `id="app"`) && !strings.Contains(dbody, "totalContacts") {
		t.Errorf("dashboard missing Inertia marker or data, body: %s", dbody)
	}

	logoutForm := url.Values{}
	logoutForm.Set("csrf_token", csrfToken)
	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(logoutForm.Encode()))
	logoutReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	logoutReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	logoutReq.AddCookie(sessionCookie)
	logoutRR := httptest.NewRecorder()
	h.ServeHTTP(logoutRR, logoutReq)
	if logoutRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /logout status = %d, want 303", logoutRR.Code)
	}
	if logoutRR.Header().Get("Location") != "/login" {
		t.Errorf("logout Location = %q, want /login", logoutRR.Header().Get("Location"))
	}

	dashAfterLogout := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashAfterLogout.AddCookie(sessionCookie)
	dashAfterRR := httptest.NewRecorder()
	h.ServeHTTP(dashAfterRR, dashAfterLogout)
	if dashAfterRR.Code != http.StatusSeeOther {
		t.Errorf("GET /dashboard after logout status = %d, want 303", dashAfterRR.Code)
	}
}

func TestApp_PasswordResetFlow(t *testing.T) {
	root := projectRoot(t)
	catalog := i18n.DefaultCatalog()
	renderer, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), catalog)
	if err != nil {
		t.Fatal(err)
	}

	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	cfg := cais.Config{Port: ":0", DBPath: ":memory:", Env: "development", AppURL: "http://localhost:8080"}
	a, err := New(cfg, Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: filepath.Join(root, "web", "static"),
		Site:      meta.SiteFrom("Cais", ""),
		Catalog:   catalog,
	})
	if err != nil {
		t.Fatal(err)
	}
	h := a.Handler()

	getForgot := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	forgotRR := httptest.NewRecorder()
	h.ServeHTTP(forgotRR, getForgot)
	csrfToken := csrfTokenFromResponse(t, forgotRR.Result())

	forgotForm := url.Values{}
	forgotForm.Set("email", "demo@example.com")
	forgotForm.Set("csrf_token", csrfToken)
	postForgot := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(forgotForm.Encode()))
	postForgot.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postForgot.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	forgotPostRR := httptest.NewRecorder()
	h.ServeHTTP(forgotPostRR, postForgot)
	if forgotPostRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /forgot-password status = %d, want 303", forgotPostRR.Code)
	}

	var resetToken string
	if err := s.DB().QueryRow("SELECT token FROM password_reset_tokens LIMIT 1").Scan(&resetToken); err != nil {
		t.Fatalf("reset token missing: %v", err)
	}

	getReset := httptest.NewRequest(http.MethodGet, "/reset-password?token="+resetToken, nil)
	resetRR := httptest.NewRecorder()
	h.ServeHTTP(resetRR, getReset)
	csrfToken = csrfTokenFromResponse(t, resetRR.Result())

	resetForm := url.Values{}
	resetForm.Set("token", resetToken)
	resetForm.Set("password", "new-password-123")
	resetForm.Set("password_confirmation", "new-password-123")
	resetForm.Set("csrf_token", csrfToken)
	postReset := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(resetForm.Encode()))
	postReset.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReset.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	resetPostRR := httptest.NewRecorder()
	h.ServeHTTP(resetPostRR, postReset)
	if resetPostRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /reset-password status = %d, want 303, body: %s", resetPostRR.Code, resetPostRR.Body.String())
	}

	getLogin := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRR := httptest.NewRecorder()
	h.ServeHTTP(loginRR, getLogin)
	csrfToken = csrfTokenFromResponse(t, loginRR.Result())

	loginForm := url.Values{}
	loginForm.Set("email", "demo@example.com")
	loginForm.Set("password", "new-password-123")
	loginForm.Set("csrf_token", csrfToken)
	postLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginForm.Encode()))
	postLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postLogin.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	loginPostRR := httptest.NewRecorder()
	h.ServeHTTP(loginPostRR, postLogin)
	if loginPostRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /login with new password status = %d, want 303", loginPostRR.Code)
	}
}

func TestApp_SignUpFlow_registersAndSignsIn(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	getSignup := httptest.NewRequest(http.MethodGet, "/signup", nil)
	signupRR := httptest.NewRecorder()
	h.ServeHTTP(signupRR, getSignup)
	csrfToken := csrfTokenFromResponse(t, signupRR.Result())

	form := url.Values{}
	form.Set("email", "newuser@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	form.Set("csrf_token", csrfToken)
	postSignup := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	postSignup.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postSignup.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	postSignupRR := httptest.NewRecorder()
	h.ServeHTTP(postSignupRR, postSignup)
	if postSignupRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /signup status = %d, want 303", postSignupRR.Code)
	}

	dashReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	for _, c := range postSignupRR.Result().Cookies() {
		dashReq.AddCookie(c)
	}
	dashRR := httptest.NewRecorder()
	h.ServeHTTP(dashRR, dashReq)
	if dashRR.Code != http.StatusOK {
		t.Fatalf("GET /dashboard status = %d, want 200", dashRR.Code)
	}
	dbody := dashRR.Body.String()
	if !strings.Contains(dbody, `id="app"`) && !strings.Contains(dbody, "totalContacts") {
		t.Errorf("dashboard missing Inertia marker or data, body: %s", dbody)
	}
}

func TestApp_Login_Inertia_TDD(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	// GET /login X-Inertia
	getReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	getReq.Header.Set("X-Inertia", "true")
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET login inertia status=%d", getRR.Code)
	}
	var gp map[string]any
	if err := json.Unmarshal(getRR.Body.Bytes(), &gp); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if gp["component"] != "Login" {
		t.Errorf("want Login component, got %v", gp["component"])
	}

	// POST bad creds with X-Inertia + csrf -> error in props
	getTok := httptest.NewRequest(http.MethodGet, "/login", nil)
	getTokRR := httptest.NewRecorder()
	h.ServeHTTP(getTokRR, getTok)
	tok := csrfTokenFromResponse(t, getTokRR.Result())

	bad := url.Values{"email": {"no@ex.com"}, "password": {"wrong"}, "csrf_token": {tok}}
	postBad := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(bad.Encode()))
	postBad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postBad.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: tok})
	postBad.Header.Set("X-Inertia", "true")
	pbRR := httptest.NewRecorder()
	h.ServeHTTP(pbRR, postBad)
	if pbRR.Code != http.StatusOK {
		t.Fatalf("bad login inertia status=%d body=%s", pbRR.Code, pbRR.Body.String())
	}
	var bp map[string]any
	if err := json.Unmarshal(pbRR.Body.Bytes(), &bp); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if bp["component"] != "Login" {
		t.Errorf("bad login should re-render Login, got %v", bp["component"])
	}
	lprops, ok := bp["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", bp)
	}
	lerrs, ok := lprops["errors"].(map[string]any)
	if !ok || lerrs["email"] == nil {
		t.Errorf("props.errors missing email: %v", lprops)
	}
}

// TestApp_StaticBuildMainJS serves the Vite-built bundle through the real app handler.
func TestApp_Dashboard_Inertia_TDD(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	getLogin := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRR := httptest.NewRecorder()
	h.ServeHTTP(loginRR, getLogin)
	csrfToken := csrfTokenFromResponse(t, loginRR.Result())

	loginForm := url.Values{}
	loginForm.Set("email", "demo@example.com")
	loginForm.Set("password", "password")
	loginForm.Set("csrf_token", csrfToken)
	postLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginForm.Encode()))
	postLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postLogin.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	loginPostRR := httptest.NewRecorder()
	h.ServeHTTP(loginPostRR, postLogin)
	if loginPostRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /login status = %d, want 303", loginPostRR.Code)
	}

	dashReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashReq.Header.Set("X-Inertia", "true")
	for _, c := range loginPostRR.Result().Cookies() {
		dashReq.AddCookie(c)
	}
	dashRR := httptest.NewRecorder()
	h.ServeHTTP(dashRR, dashReq)

	if dashRR.Code != http.StatusOK {
		t.Fatalf("GET /dashboard X-Inertia status = %d, body=%s", dashRR.Code, dashRR.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(dashRR.Body.Bytes(), &payload); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if payload["component"] != "Dashboard" {
		t.Errorf("component = %v, want Dashboard", payload["component"])
	}
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	if _, ok := props["totalContacts"]; !ok {
		t.Errorf("props missing totalContacts: %v", props)
	}
}

// TestApp_AuthPages_Inertia_TDD covers signup, forgot-password, and reset-password Inertia responses.
func TestApp_AuthPages_Inertia_TDD(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	// Signup GET
	signupReq := httptest.NewRequest(http.MethodGet, "/signup", nil)
	signupReq.Header.Set("X-Inertia", "true")
	signupRR := httptest.NewRecorder()
	h.ServeHTTP(signupRR, signupReq)
	var signupPayload map[string]any
	if err := json.Unmarshal(signupRR.Body.Bytes(), &signupPayload); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if signupPayload["component"] != "Signup" {
		t.Errorf("signup component = %v, want Signup", signupPayload["component"])
	}

	// Signup POST validation error
	getSignup := httptest.NewRequest(http.MethodGet, "/signup", nil)
	getSignupRR := httptest.NewRecorder()
	h.ServeHTTP(getSignupRR, getSignup)
	csrfToken := csrfTokenFromResponse(t, getSignupRR.Result())
	badSignup := url.Values{"email": {"bad"}, "password": {"short"}, "password_confirmation": {"x"}, "csrf_token": {csrfToken}}
	postSignup := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(badSignup.Encode()))
	postSignup.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postSignup.Header.Set("X-Inertia", "true")
	postSignup.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	postSignupRR := httptest.NewRecorder()
	h.ServeHTTP(postSignupRR, postSignup)
	var signupPost map[string]any
	if err := json.Unmarshal(postSignupRR.Body.Bytes(), &signupPost); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if signupPost["component"] != "Signup" {
		t.Errorf("bad signup component = %v", signupPost["component"])
	}
	if p, ok := signupPost["props"].(map[string]any); ok {
		if e, ok := p["errors"].(map[string]any); !ok || len(e) == 0 {
			t.Errorf("bad signup missing errors in props: %v", p)
		}
	} else {
		t.Errorf("bad signup missing props")
	}

	// Forgot password GET
	forgotReq := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	forgotReq.Header.Set("X-Inertia", "true")
	forgotRR := httptest.NewRecorder()
	h.ServeHTTP(forgotRR, forgotReq)
	var forgotPayload map[string]any
	if err := json.Unmarshal(forgotRR.Body.Bytes(), &forgotPayload); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if forgotPayload["component"] != "ForgotPassword" {
		t.Errorf("forgot component = %v, want ForgotPassword", forgotPayload["component"])
	}

	// Forgot password POST validation
	getForgot := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	getForgotRR := httptest.NewRecorder()
	h.ServeHTTP(getForgotRR, getForgot)
	csrfToken = csrfTokenFromResponse(t, getForgotRR.Result())
	badForgot := url.Values{"email": {"not-email"}, "csrf_token": {csrfToken}}
	postForgot := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(badForgot.Encode()))
	postForgot.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postForgot.Header.Set("X-Inertia", "true")
	postForgot.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	postForgotRR := httptest.NewRecorder()
	h.ServeHTTP(postForgotRR, postForgot)
	var forgotPost map[string]any
	if err := json.Unmarshal(postForgotRR.Body.Bytes(), &forgotPost); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if p, ok := forgotPost["props"].(map[string]any); ok {
		if e, ok := p["errors"].(map[string]any); !ok || len(e) == 0 {
			t.Errorf("forgot POST missing errors: %v", p)
		}
	} else {
		t.Errorf("forgot POST missing props")
	}

	// Reset password GET with invalid token
	resetReq := httptest.NewRequest(http.MethodGet, "/reset-password?token=bad", nil)
	resetReq.Header.Set("X-Inertia", "true")
	resetRR := httptest.NewRecorder()
	h.ServeHTTP(resetRR, resetReq)
	var resetPayload map[string]any
	if err := json.Unmarshal(resetRR.Body.Bytes(), &resetPayload); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if resetPayload["component"] != "ResetPassword" {
		t.Errorf("reset component = %v, want ResetPassword", resetPayload["component"])
	}
	rprops, ok := resetPayload["props"].(map[string]any)
	if !ok {
		t.Fatalf("reset missing props: %v", resetPayload)
	}
	rerrs, ok := rprops["errors"].(map[string]any)
	if !ok || rerrs["token"] == nil {
		t.Errorf("reset props.errors missing token: %v", rprops)
	}
}
