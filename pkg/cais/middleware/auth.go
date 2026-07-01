package middleware

import (
	"net/http"
	"os"
	"strings"
)

// TokenAuth protects routes with a bearer token or ?token= query param.
// Set ADMIN_TOKEN env var. If empty, middleware is a no-op (dev mode).
func TokenAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := os.Getenv("ADMIN_TOKEN")
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if extractToken(r) != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Protect wraps a handler with TokenAuth.
func Protect(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		TokenAuth(http.HandlerFunc(h)).ServeHTTP(w, r)
	}
}

func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return r.URL.Query().Get("token")
}
