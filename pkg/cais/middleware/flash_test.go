package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/flash"
)

func TestFlash_PutsMessageInContext(t *testing.T) {
	rr := httptest.NewRecorder()
	flash.Set(rr, "notice", "Bem-vindo!", false)

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()

	var flashCookie *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == flash.CookieName {
			flashCookie = c
			break
		}
	}
	if flashCookie == nil {
		t.Fatal("flash cookie not set")
	}

	var got flash.Message
	var gotOK bool
	h := Flash(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, gotOK = FlashMessage(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.AddCookie(flashCookie)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !gotOK {
		t.Fatal("FlashMessage returned false")
	}
	if got.Kind != "notice" || got.Message != "Bem-vindo!" {
		t.Errorf("message = %+v, want notice/Bem-vindo!", got)
	}

	clearRes := rec.Result()
	defer func() { _ = clearRes.Body.Close() }()
	var cleared bool
	for _, c := range clearRes.Cookies() {
		if c.Name == flash.CookieName && c.MaxAge == -1 {
			cleared = true
			break
		}
	}
	if !cleared {
		t.Error("flash cookie not cleared on response")
	}
}

func TestFlash_NoCookie_PassesThrough(t *testing.T) {
	called := false
	h := Flash(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := FlashMessage(r); ok {
			t.Error("expected no flash message")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Error("handler not called")
	}
}
