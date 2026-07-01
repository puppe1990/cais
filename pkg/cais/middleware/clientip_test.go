package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestClientIP_untrustedIgnoresXFF(t *testing.T) {
	cfg := cais.Config{} // no trusted proxies
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	req.RemoteAddr = "198.51.100.1:1234"
	if got := ClientIP(req, cfg); got != "198.51.100.1" {
		t.Errorf("got %q", got)
	}
}

func TestClientIP_trustedUsesXFF(t *testing.T) {
	cfg := cais.Config{TrustedProxies: []string{"127.0.0.1"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")
	req.RemoteAddr = "127.0.0.1:9999"
	if got := ClientIP(req, cfg); got != "203.0.113.50" {
		t.Errorf("got %q", got)
	}
}

func TestClientIP_untrustedIgnoresXRealIP(t *testing.T) {
	cfg := cais.Config{}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.5")
	req.RemoteAddr = "198.51.100.1:1234"
	if got := ClientIP(req, cfg); got != "198.51.100.1" {
		t.Errorf("got %q", got)
	}
}

func TestClientIP_trustedFallsBackToRemoteAddr(t *testing.T) {
	cfg := cais.Config{TrustedProxies: []string{"127.0.0.1"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	if got := ClientIP(req, cfg); got != "127.0.0.1" {
		t.Errorf("got %q", got)
	}
}

func TestClientIP_trustedUsesXRealIPWhenNoXFF(t *testing.T) {
	cfg := cais.Config{TrustedProxies: []string{"127.0.0.1"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.77")
	req.RemoteAddr = "127.0.0.1:9999"
	if got := ClientIP(req, cfg); got != "203.0.113.77" {
		t.Errorf("got %q, want 203.0.113.77", got)
	}
}

func TestClientIP_trustedCIDR(t *testing.T) {
	cfg := cais.Config{TrustedProxies: []string{"10.0.0.0/8"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	req.RemoteAddr = "10.1.2.3:8080"
	if got := ClientIP(req, cfg); got != "203.0.113.50" {
		t.Errorf("got %q, want 203.0.113.50", got)
	}
}

func TestClientIP_untrustedCIDR(t *testing.T) {
	cfg := cais.Config{TrustedProxies: []string{"10.0.0.0/8"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	req.RemoteAddr = "198.51.100.1:8080"
	if got := ClientIP(req, cfg); got != "198.51.100.1" {
		t.Errorf("got %q, want 198.51.100.1", got)
	}
}
