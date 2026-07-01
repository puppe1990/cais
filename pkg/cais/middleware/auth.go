package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
)

// AdminAuth protects routes with a Bearer token from cfg.AdminToken.
// In development with an empty token, requests pass through.
// In production with an empty token, all requests are rejected.
func AdminAuth(cfg cais.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.AdminToken == "" {
				if cfg.Env == "production" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if !bearerMatches(r, cfg.AdminToken) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Protect wraps a handler with AdminAuth.
func Protect(cfg cais.Config, h http.HandlerFunc) http.HandlerFunc {
	auth := AdminAuth(cfg)
	return func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(h)).ServeHTTP(w, r)
	}
}

// TokenAuth is deprecated; use AdminAuth(cfg) instead.
//
// Deprecated: TokenAuth loads config at call time and will be removed in a future release.
func TokenAuth(next http.Handler) http.Handler {
	return AdminAuth(cais.Load())(next)
}

func bearerMatches(r *http.Request, want string) bool {
	got := bearerToken(r)
	if got == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
