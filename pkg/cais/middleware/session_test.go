package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestLoadSession_SetsUserID(t *testing.T) {
	store := session.NewMemoryStore()
	token, err := store.Create(99)
	if err != nil {
		t.Fatal(err)
	}

	var gotID int64
	var gotOK bool
	h := LoadSession(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, gotOK = session.UserID(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: session.DefaultCookieName, Value: token})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !gotOK || gotID != 99 {
		t.Fatalf("user = (%d, %v), want (99, true)", gotID, gotOK)
	}
}

func TestLoadSession_NoCookie_PassesThrough(t *testing.T) {
	store := session.NewMemoryStore()
	called := false
	h := LoadSession(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := session.UserID(r); ok {
			t.Error("expected no user without cookie")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Error("handler not called")
	}
}

func TestRequireAuth_RedirectsWhenAnonymous(t *testing.T) {
	h := RequireAuth("/login")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not run")
	}))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if loc != "/login" {
		t.Errorf("Location = %q, want /login", loc)
	}
}

func TestRequireAuth_AllowsAuthenticatedUser(t *testing.T) {
	h := RequireAuth("/login")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequireAuth_HTMX_SetsRedirectHeader(t *testing.T) {
	h := RequireAuth("/login")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not run")
	}))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
	if got := rr.Header().Get("HX-Redirect"); got != "/login" {
		t.Errorf("HX-Redirect = %q, want /login", got)
	}
}
