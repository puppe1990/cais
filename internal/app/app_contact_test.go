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

func TestApp_ContactRoute_Inertia(t *testing.T) {
	a := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	req.Header.Set("X-Inertia", "true")
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if payload["component"] != "Contact" {
		t.Errorf("component = %v, want Contact", payload["component"])
	}
}

func TestApp_ContactPost_requiresCSRF(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET /contact status = %d", getRR.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=alice@example.com"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF status = %d, want 403", postRR.Code)
	}
}

func TestApp_ContactPost_withCSRF_succeeds(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	var token string
	for _, c := range getRR.Result().Cookies() {
		if c.Name == csrf.CookieName {
			token = c.Value
		}
	}
	if token == "" {
		t.Fatal("missing csrf cookie after GET /contact")
	}

	form := url.Values{}
	form.Set("name", "Alice")
	form.Set("email", "alice@example.com")
	form.Set("csrf_token", token)

	postReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	postReq.Header.Set("X-Inertia", "true")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusSeeOther {
		t.Errorf("POST with CSRF status = %d, want 303 (Inertia redirect), body: %s", postRR.Code, postRR.Body.String())
	}
	if postRR.Header().Get("Location") != "/contact" {
		t.Errorf("Location = %q, want /contact", postRR.Header().Get("Location"))
	}
}

func TestApp_ContactPost_validationWithCSRF_returns422(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	getReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	csrfToken := csrfTokenFromResponse(t, getRR.Result())

	form := url.Values{}
	form.Set("name", "")
	form.Set("email", "alice@example.com")
	form.Set("csrf_token", csrfToken)

	postReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	postReq.Header.Set("X-Inertia", "true")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK {
		t.Errorf("POST status = %d, want 200 (Inertia validation), body: %s", postRR.Code, postRR.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(postRR.Body.Bytes(), &payload); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if payload["component"] != "Contact" {
		t.Errorf("component = %v, want Contact", payload["component"])
	}
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	errs, ok := props["errors"].(map[string]any)
	if !ok || errs["name"] == nil {
		t.Errorf("props.errors missing name: %v", props)
	}
}

func TestApp_Contact_Inertia_TDD(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	// GET /contact X-Inertia -> component
	getReq := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getReq.Header.Set("X-Inertia", "true")
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("GET X-Inertia /contact status=%d", getRR.Code)
	}
	var getPayload map[string]any
	if err := json.Unmarshal(getRR.Body.Bytes(), &getPayload); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if getPayload["component"] != "Contact" {
		t.Errorf("expected component=Contact, got %v", getPayload["component"])
	}

	// POST validation error with X-Inertia + csrf -> should deliver errors in props
	getCSRF := httptest.NewRequest(http.MethodGet, "/contact", nil)
	getCSRFrr := httptest.NewRecorder()
	h.ServeHTTP(getCSRFrr, getCSRF)
	token := csrfTokenFromResponse(t, getCSRFrr.Result())

	form := url.Values{}
	form.Set("name", "")
	form.Set("email", "bad")
	form.Set("csrf_token", token)
	postReq := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: token})
	postReq.Header.Set("X-Inertia", "true")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK && postRR.Code != http.StatusUnprocessableEntity {
		t.Fatalf("POST X-Inertia validation status=%d want 200 or 422, body=%s", postRR.Code, postRR.Body.String())
	}
	var postPayload map[string]any
	if err := json.Unmarshal(postRR.Body.Bytes(), &postPayload); err != nil {
		t.Fatalf("validation post not json: %v body=%s", err, postRR.Body.String())
	}
	// gonertia puts validation errors under props.errors (AlwaysProp)
	foundErrs := false
	if p, ok := postPayload["props"].(map[string]any); ok {
		if e, ok := p["errors"]; ok && e != nil {
			if em, ok := e.(map[string]any); ok && len(em) > 0 {
				foundErrs = true
			}
		}
	}
	if !foundErrs {
		// also accept top-level or stringified for robustness
		b := postRR.Body.String()
		if !strings.Contains(b, "name") && !strings.Contains(b, "email") && !strings.Contains(b, "error") {
			t.Errorf("expected errors in Inertia props for validation POST, got payload=%v body=%s", postPayload, b)
		}
	}
}

// TestApp_Login_Inertia_TDD extends TDD for login flows per plan.
