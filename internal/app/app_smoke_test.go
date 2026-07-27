package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/csrf"
)

func TestApp_Smoke_contactInertia_loginDashboardLogout(t *testing.T) {
	a := setupTestAppDev(t)
	h := a.Handler()

	getContact := httptest.NewRequest(http.MethodGet, "/contact", nil)
	contactRR := httptest.NewRecorder()
	h.ServeHTTP(contactRR, getContact)
	csrfToken := csrfTokenFromResponse(t, contactRR.Result())

	contactForm := url.Values{}
	contactForm.Set("name", "Smoke Test")
	contactForm.Set("email", "smoke@example.com")
	contactForm.Set("csrf_token", csrfToken)
	postContact := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(contactForm.Encode()))
	postContact.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postContact.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	postContact.Header.Set("X-Inertia", "true")
	contactPostRR := httptest.NewRecorder()
	h.ServeHTTP(contactPostRR, postContact)
	if contactPostRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /contact status = %d, want 303", contactPostRR.Code)
	}

	followContact := httptest.NewRequest(http.MethodGet, "/contact", nil)
	followContact.Header.Set("X-Inertia", "true")
	for _, c := range contactPostRR.Result().Cookies() {
		followContact.AddCookie(c)
	}
	followRR := httptest.NewRecorder()
	h.ServeHTTP(followRR, followContact)
	var contactPayload map[string]any
	if err := json.Unmarshal(followRR.Body.Bytes(), &contactPayload); err != nil {
		t.Fatalf("follow GET not json: %v", err)
	}
	if contactPayload["component"] != "Contact" {
		t.Errorf("component = %v, want Contact", contactPayload["component"])
	}
	cprops, ok := contactPayload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", contactPayload)
	}
	flash, ok := cprops["flash"].(map[string]any)
	if !ok || flash["success"] == nil {
		t.Errorf("props.flash missing success after POST redirect: %v", cprops)
	}

	getLogin := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRR := httptest.NewRecorder()
	h.ServeHTTP(loginRR, getLogin)
	csrfToken = csrfTokenFromResponse(t, loginRR.Result())

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
		t.Fatalf("GET /dashboard status = %d, want 200", dashRR.Code)
	}
	var dashPayload map[string]any
	if err := json.Unmarshal(dashRR.Body.Bytes(), &dashPayload); err != nil {
		t.Fatalf("dashboard not json: %v", err)
	}
	if dashPayload["component"] != "Dashboard" {
		t.Errorf("component = %v, want Dashboard", dashPayload["component"])
	}

	logoutForm := url.Values{}
	logoutForm.Set("csrf_token", csrfToken)
	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(logoutForm.Encode()))
	logoutReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	logoutReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	logoutReq.AddCookie(sessionCookieFromResponse(t, loginPostRR.Result()))
	logoutRR := httptest.NewRecorder()
	h.ServeHTTP(logoutRR, logoutReq)
	if logoutRR.Code != http.StatusSeeOther {
		t.Fatalf("POST /logout status = %d, want 303", logoutRR.Code)
	}
}

func TestApp_HomeRoute_Inertia(t *testing.T) {
	a := setupTestApp(t)

	// Ordinary request: must yield Inertia root HTML shell
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	hasInertiaMarker := strings.Contains(body, `id="app"`) ||
		strings.Contains(body, "data-page") ||
		strings.Contains(body, "{{ .inertia }}") ||
		strings.Contains(body, "inertia")
	if !hasInertiaMarker {
		t.Errorf("expected Inertia root shell markers (id=app or data-page or .inertia), got body: %s", body)
	}

	// X-Inertia request: must yield protocol JSON with component + props
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Inertia", "true")
	rr2 := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("X-Inertia status = %d, want 200", rr2.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rr2.Body.Bytes(), &payload); err != nil {
		t.Fatalf("X-Inertia body not JSON: %v, body=%s", err, rr2.Body.String())
	}
	if _, ok := payload["component"]; !ok {
		t.Errorf("X-Inertia JSON missing 'component' key, got: %v", payload)
	}
	if _, ok := payload["props"]; !ok {
		t.Errorf("X-Inertia JSON missing 'props' key, got: %v", payload)
	}
}

// TestApp_Contact_Inertia_TDD extends TDD to contact: written first as failing test
// asserting Inertia component + error props for validation under real entrypoint.
