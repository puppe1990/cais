package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignIn_SetsCookieAndSession(t *testing.T) {
	store := NewMemoryStore()
	rr := httptest.NewRecorder()

	if err := SignIn(rr, store, httptest.NewRequest(http.MethodPost, "/login", nil), 5, CookieOptions{}); err != nil {
		t.Fatal(err)
	}

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}

	id, ok := store.Get(cookies[0].Value)
	if !ok || id != 5 {
		t.Fatalf("store.Get() = (%d, %v), want (5, true)", id, ok)
	}
}

func TestSignIn_rotatesExistingSession(t *testing.T) {
	store := NewMemoryStore()
	oldToken, err := store.Create(9)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: oldToken})
	rr := httptest.NewRecorder()

	if err := SignIn(rr, store, req, 9, CookieOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, ok := store.Get(oldToken); ok {
		t.Fatal("old session token should be invalidated on login")
	}

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()
	newToken := res.Cookies()[0].Value
	if newToken == oldToken {
		t.Fatal("expected a new session token after login")
	}
	if id, ok := store.Get(newToken); !ok || id != 9 {
		t.Fatalf("new token lookup = (%d, %v), want (9, true)", id, ok)
	}
}

func TestSignOut_ClearsCookieAndSession(t *testing.T) {
	store := NewMemoryStore()
	token, err := store.Create(3)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: token})
	rr := httptest.NewRecorder()

	SignOut(rr, store, req, CookieOptions{})

	if _, ok := store.Get(token); ok {
		t.Error("session should be deleted")
	}
	res := rr.Result()
	defer func() { _ = res.Body.Close() }()
	if len(res.Cookies()) != 1 || res.Cookies()[0].MaxAge != -1 {
		t.Error("expected cleared cookie")
	}
}
