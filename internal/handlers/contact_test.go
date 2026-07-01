package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matheuspuppe/cais/internal/store"
)

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestContactHandler_Get_ReturnsForm(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t))

	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "contact-form") {
		t.Errorf("body missing form, got: %s", rr.Body.String())
	}
}

func TestContactHandler_Post_InvalidEmail_Returns422(t *testing.T) {
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t))

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
	h := NewContactHandler(setupTestRenderer(t), setupTestStore(t))

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("expected partial HTML, got full page")
	}
	if !strings.Contains(body, "email") {
		t.Errorf("body missing error message, got: %s", body)
	}
}

func TestContactHandler_Post_Valid_SavesAndReturnsSuccess(t *testing.T) {
	s := setupTestStore(t)
	h := NewContactHandler(setupTestRenderer(t), s)

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=Alice&email=alice@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "sucesso") {
		t.Errorf("body missing success message, got: %s", rr.Body.String())
	}
}
