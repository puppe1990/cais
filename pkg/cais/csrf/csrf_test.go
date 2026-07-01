package csrf

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateToken_isUnique(t *testing.T) {
	a, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected unique tokens")
	}
	if len(a) < 32 {
		t.Errorf("token too short: %q", a)
	}
}

func TestSubmittedToken_prefersHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("csrf_token=from-form"))
	req.Header.Set("X-CSRF-Token", "from-header")
	_ = req.ParseForm()

	got := SubmittedToken(req)
	if got != "from-header" {
		t.Errorf("SubmittedToken = %q, want from-header", got)
	}
}

func TestSubmittedToken_readsFormField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("csrf_token=from-form"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	got := SubmittedToken(req)
	if got != "from-form" {
		t.Errorf("SubmittedToken = %q, want from-form", got)
	}
}

func TestValid_matchingCookieAndSubmitted(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "abc123"})
	req.Header.Set("X-CSRF-Token", "abc123")

	if !Valid(req) {
		t.Fatal("expected valid CSRF")
	}
}

func TestValid_rejectsMissingCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-CSRF-Token", "abc123")

	if Valid(req) {
		t.Fatal("expected invalid without cookie")
	}
}

func TestValid_rejectsMismatch(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "abc123"})
	req.Header.Set("X-CSRF-Token", "wrong")

	if Valid(req) {
		t.Fatal("expected invalid on mismatch")
	}
}

func TestSetCookie_andTokenFromRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	token, err := EnsureToken(rr, req)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected token")
	}

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()
	cookies := res.Cookies()
	if len(cookies) == 0 || cookies[0].Name != CookieName {
		t.Fatalf("cookie not set: %+v", cookies)
	}

	req2 := req.WithContext(WithToken(req.Context(), token))
	if got := TokenFromRequest(req2); got != token {
		t.Errorf("TokenFromRequest = %q, want %q", got, token)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	if got := TokenFromRequest(req3); got != token {
		t.Errorf("TokenFromRequest from cookie = %q, want %q", got, token)
	}
}

func TestFieldHTML_escapesValue(t *testing.T) {
	got := FieldHTML(`"><script>`)
	if strings.Contains(got, "<script>") {
		t.Errorf("FieldHTML not escaped: %s", got)
	}
	if !strings.Contains(got, `name="csrf_token"`) {
		t.Errorf("FieldHTML missing field name: %s", got)
	}
}
