package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseFormOrJSON_jsonBody(t *testing.T) {
	body := `{"email":"demo@example.com","password":"secret","remember":true,"count":3}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if err := ParseFormOrJSON(req); err != nil {
		t.Fatal(err)
	}
	if got := req.FormValue("email"); got != "demo@example.com" {
		t.Errorf("email = %q, want demo@example.com", got)
	}
	if got := req.FormValue("password"); got != "secret" {
		t.Errorf("password = %q, want secret", got)
	}
	if got := req.FormValue("remember"); got != "true" {
		t.Errorf("remember = %q, want true", got)
	}
	if got := req.FormValue("count"); got != "3" {
		t.Errorf("count = %q, want 3", got)
	}
}

func TestParseFormOrJSON_jsonWithCharset(t *testing.T) {
	body := `{"name":"Ada"}`
	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if err := ParseFormOrJSON(req); err != nil {
		t.Fatal(err)
	}
	if got := req.FormValue("name"); got != "Ada" {
		t.Errorf("name = %q, want Ada", got)
	}
}

func TestParseFormOrJSON_urlencoded(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=a%40b.com&password=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := ParseFormOrJSON(req); err != nil {
		t.Fatal(err)
	}
	if got := req.FormValue("email"); got != "a@b.com" {
		t.Errorf("email = %q, want a@b.com", got)
	}
	if got := req.FormValue("password"); got != "x" {
		t.Errorf("password = %q, want x", got)
	}
}

func TestParseFormOrJSON_afterMiddlewareParseForm(t *testing.T) {
	// CSRF middleware calls ParseForm before the handler; JSON body must still parse.
	body := `{"email":"after@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	if got := req.FormValue("email"); got != "" {
		t.Fatalf("precondition: ParseForm should leave JSON unread, got %q", got)
	}

	if err := ParseFormOrJSON(req); err != nil {
		t.Fatal(err)
	}
	if got := req.FormValue("email"); got != "after@example.com" {
		t.Errorf("email = %q, want after@example.com", got)
	}
}

func TestParseFormOrJSON_invalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	if err := ParseFormOrJSON(req); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFormTruthy(t *testing.T) {
	for _, v := range []string{"on", "true", "TRUE", "1", "yes", " Yes "} {
		if !FormTruthy(v) {
			t.Errorf("FormTruthy(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"", "false", "0", "off", "no"} {
		if FormTruthy(v) {
			t.Errorf("FormTruthy(%q) = true, want false", v)
		}
	}
}
