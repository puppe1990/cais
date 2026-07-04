package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestSecurityHeaders_production(t *testing.T) {
	cfg := cais.Config{Env: "production", AppURL: "https://app.example.com"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, key := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Permissions-Policy",
		"Content-Security-Policy",
		"Strict-Transport-Security",
	} {
		if rr.Header().Get(key) == "" {
			t.Errorf("missing header %s", key)
		}
	}
}

func TestSecurityHeaders_customPolicy(t *testing.T) {
	cfg := cais.Config{
		Env:               "development",
		PermissionsPolicy: "camera=(self), geolocation=(self)",
		CSPStyleSrc:       "https://fonts.googleapis.com",
	}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rr.Header().Get("Permissions-Policy"); got != "camera=(self), geolocation=(self)" {
		t.Errorf("Permissions-Policy = %q", got)
	}
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://fonts.googleapis.com") {
		t.Errorf("CSP missing style src extra: %q", csp)
	}
}

func TestSecurityHeaders_mediaSrc(t *testing.T) {
	cfg := cais.Config{
		Env:         "development",
		CSPMediaSrc: "blob:",
	}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "media-src 'self' blob:") {
		t.Errorf("CSP missing media-src blob:, got %q", csp)
	}
}

func TestSecurityHeaders_development_allowsCamera(t *testing.T) {
	cfg := cais.Config{Env: "development", PermissionsPolicy: "camera=(self), microphone=(), geolocation=()", CSPMediaSrc: "blob:"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rr.Header().Get("Permissions-Policy"); !strings.Contains(got, "camera=(self)") {
		t.Errorf("Permissions-Policy = %q, want camera=(self)", got)
	}
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "media-src 'self' blob:") {
		t.Errorf("CSP = %q", csp)
	}
}

func TestSecurityHeaders_development_noHSTS(t *testing.T) {
	cfg := cais.Config{Env: "development"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS should not be set in development")
	}
}
