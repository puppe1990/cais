package middleware

import (
	"fmt"
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
)

// SecurityHeaders sets baseline headers. CSP allows 'unsafe-inline' for HTMX and inline layout scripts.
// Nonce-based CSP needs asset bundling changes — see README security section.
func SecurityHeaders(cfg cais.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			policy := cfg.PermissionsPolicy
			if policy == "" {
				policy = "camera=(), microphone=(), geolocation=()"
			}
			w.Header().Set("Permissions-Policy", policy)
			styleSrc := "'self' 'unsafe-inline'"
			if cfg.CSPStyleSrc != "" {
				styleSrc += " " + cfg.CSPStyleSrc
			}
			connectSrc := "'self'"
			if cfg.CSPConnectSrc != "" {
				connectSrc += " " + cfg.CSPConnectSrc
			}
			mediaSrc := "'self'"
			if cfg.CSPMediaSrc != "" {
				mediaSrc += " " + cfg.CSPMediaSrc
			}
			imgSrc := "'self' data:"
			if cfg.CSPImgSrc != "" {
				imgSrc += " " + cfg.CSPImgSrc
			}
			w.Header().Set("Content-Security-Policy", fmt.Sprintf(
				"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src %s; img-src %s; connect-src %s; media-src %s; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
				styleSrc, imgSrc, connectSrc, mediaSrc,
			))
			if cfg.Env == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}
