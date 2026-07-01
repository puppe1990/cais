package middleware

import (
	"net/http"
	"net/http/httptest"
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
		"Content-Security-Policy",
		"Strict-Transport-Security",
	} {
		if rr.Header().Get(key) == "" {
			t.Errorf("missing header %s", key)
		}
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
