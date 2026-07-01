package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCookie_AndTokenFromRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	SetCookie(rr, "abc123", CookieOptions{Secure: false})

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()

	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	c := cookies[0]
	if c.Name != DefaultCookieName {
		t.Errorf("name = %q, want %q", c.Name, DefaultCookieName)
	}
	if c.Value != "abc123" {
		t.Errorf("value = %q, want abc123", c.Value)
	}
	if !c.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(c)
	if got := TokenFromRequest(req); got != "abc123" {
		t.Errorf("TokenFromRequest() = %q, want abc123", got)
	}
}

func TestClearCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	ClearCookie(rr)

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()

	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	if cookies[0].MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
	}
}
