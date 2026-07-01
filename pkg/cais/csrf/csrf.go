package csrf

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
)

const (
	CookieName   = "cais_csrf"
	HeaderName   = "X-CSRF-Token"
	FormField    = "csrf_token"
	MetaTag      = "csrf-token"
	CookieMaxAge = 86400 * 7
)

type ctxKey struct{}

// GenerateToken returns a random URL-safe CSRF token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("csrf token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// SubmittedToken returns the token from the request header or form field.
func SubmittedToken(r *http.Request) string {
	if v := r.Header.Get(HeaderName); v != "" {
		return v
	}
	return r.FormValue(FormField)
}

// Valid reports whether the submitted token matches the cookie (double-submit).
func Valid(r *http.Request) bool {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	submitted := SubmittedToken(r)
	if submitted == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(submitted)) == 1
}

// EnsureToken sets the CSRF cookie when missing and returns the active token.
func EnsureToken(w http.ResponseWriter, r *http.Request) (string, error) {
	if cookie, err := r.Cookie(CookieName); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})
	return token, nil
}

// WithToken stores the token in the request context.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxKey{}, token)
}

// TokenFromRequest returns the CSRF token from context or cookie.
func TokenFromRequest(r *http.Request) string {
	if token, ok := r.Context().Value(ctxKey{}).(string); ok && token != "" {
		return token
	}
	if cookie, err := r.Cookie(CookieName); err == nil {
		return cookie.Value
	}
	return ""
}

// FieldHTML returns a hidden input for HTML forms.
func FieldHTML(token string) string {
	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s" />`, FormField, template.HTMLEscapeString(token))
}

// MetaHTML returns a meta tag for HTMX requests.
func MetaHTML(token string) string {
	return fmt.Sprintf(`<meta name="%s" content="%s" />`, MetaTag, template.HTMLEscapeString(token))
}

// HTMXScript configures HTMX to send the CSRF header on every request.
func HTMXScript() string {
	return `<script>
      document.body.addEventListener("htmx:configRequest", function (evt) {
        var el = document.querySelector('meta[name="` + MetaTag + `"]');
        if (el && el.content) {
          evt.detail.headers["` + HeaderName + `"] = el.content;
        }
      });
    </script>`
}
