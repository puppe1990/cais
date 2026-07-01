package middleware

import (
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais/csrf"
)

// CSRF protects state-changing requests with a double-submit cookie token.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipCSRF(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		if isSafeMethod(r.Method) {
			token, err := csrf.EnsureToken(w, r)
			if err != nil {
				http.Error(w, "csrf token error", http.StatusInternalServerError)
				return
			}
			r = r.WithContext(csrf.WithToken(r.Context(), token))
			next.ServeHTTP(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if !csrf.Valid(r) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func skipCSRF(path string) bool {
	return path == "/health" || strings.HasPrefix(path, "/static/")
}
