package middleware

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"
)

// LoadSession reads the session cookie and attaches the user ID to the request context.
func LoadSession(store session.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := session.TokenFromRequest(r); token != "" {
				if id, ok := store.Get(token); ok {
					r = session.WithUserID(r, id)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth blocks unauthenticated requests. HTMX requests get HX-Redirect; others get 303 to loginURL.
func RequireAuth(loginURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := session.UserID(r); !ok {
				if cais.IsHTMX(r) {
					w.Header().Set("HX-Redirect", loginURL)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, loginURL, http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
