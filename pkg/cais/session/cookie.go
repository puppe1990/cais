package session

import (
	"net/http"
	"time"
)

const DefaultCookieName = "cais_session"

const defaultMaxAge = 7 * 24 * 60 * 60 // 7 days

type CookieOptions struct {
	Secure bool
}

func SetCookie(w http.ResponseWriter, token string, opts CookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     DefaultCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   defaultMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   opts.Secure,
	})
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     DefaultCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func TokenFromRequest(r *http.Request) string {
	c, err := r.Cookie(DefaultCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
