package middleware

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
)

// Flash consumes any flash cookie, clears it, and stores the message in context.
// The config drives the cookie Secure flag so deletion matches Set (#174).
func Flash(cfg cais.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if msg, ok := flash.Consume(r); ok {
				flash.Clear(w, cfg.CookieSecure())
				r = flash.WithMessage(r, msg)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// FlashMessage returns the flash message attached by the Flash middleware.
func FlashMessage(r *http.Request) (flash.Message, bool) {
	return flash.MessageFromRequest(r)
}
