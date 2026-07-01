package flash

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

const (
	CookieName   = "cais_flash"
	CookieMaxAge = 60
)

type ctxKey struct{}

// Message is a one-shot flash payload stored in a short-lived cookie.
type Message struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Set stores a flash message in the response cookie.
func Set(w http.ResponseWriter, kind, message string) {
	payload, err := json.Marshal(Message{Kind: kind, Message: message})
	if err != nil {
		return
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    encoded,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		SameSite: http.SameSiteLaxMode,
	})
}

// Consume reads the flash message from the request cookie.
func Consume(r *http.Request) (Message, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return Message{}, false
	}

	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return Message{}, false
	}

	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return Message{}, false
	}
	if msg.Message == "" {
		return Message{}, false
	}
	return msg, true
}

// Clear removes the flash cookie from the response.
func Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}

// WithMessage attaches a flash message to the request context.
func WithMessage(r *http.Request, msg Message) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, msg))
}

// MessageFromRequest returns the flash message from request context.
func MessageFromRequest(r *http.Request) (Message, bool) {
	msg, ok := r.Context().Value(ctxKey{}).(Message)
	if !ok || msg.Message == "" {
		return Message{}, false
	}
	return msg, true
}
