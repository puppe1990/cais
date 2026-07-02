package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
)

func newAuthHandlerForSignup(t *testing.T) (*AuthHandler, store.Store) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewAuthHandler(setupTestRenderer(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog())
	return h, s
}

func TestAuth_SignUpPost_createsUserAndRedirects(t *testing.T) {
	h, s := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", rr.Header().Get("Location"))
	}

	user, err := s.FindUserByEmail("signup@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Fatal("user id = 0")
	}
}

func TestAuth_SignUpPost_duplicateEmail_returnsError(t *testing.T) {
	h, _ := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("first signup status = %d, want 303", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr2 := httptest.NewRecorder()
	h.SignUpPost(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("duplicate signup status = %d, want 200", rr2.Code)
	}
	if !strings.Contains(rr2.Body.String(), "already registered") {
		t.Errorf("body missing duplicate error: %s", rr2.Body.String())
	}
}
