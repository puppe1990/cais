package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
)

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestContactHandler_Get_ReturnsForm(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "contact-form") {
		t.Errorf("body missing form, got: %s", body)
	}
	for _, needle := range []string{
		"hx-indicator",
		"hx-disabled-elt",
		"swap:150ms",
		"contact-spinner",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("body missing %q", needle)
		}
	}
}

func TestContactHandler_Post_MalformedEmail_Returns422(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=not-an-email"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(rr.Body.String(), "valid email") {
		t.Errorf("body missing validation message: %s", rr.Body.String())
	}
}

func TestContactHandler_Post_MissingName_Returns422(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(rr.Body.String(), "Name is required") {
		t.Errorf("body missing name validation: %s", rr.Body.String())
	}
}

func TestContactHandler_Post_InvalidEmail_Returns422(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestContactHandler_Post_InvalidEmail_ReturnsPartial(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("expected partial HTML, got full page")
	}
	if !strings.Contains(body, "Email is required") {
		t.Errorf("body missing error message, got: %s", body)
	}
}

func TestContactHandler_Post_Valid_SavesAndReturnsSuccess(t *testing.T) {
	s := setupTestStore(t)
	h := NewContactHandler(setupTestRenderer(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "successfully") {
		t.Errorf("body missing success message, got: %s", rr.Body.String())
	}
}
