package cais

import (
	"net/http/httptest"
	"testing"
)

func TestIsHTMX_True(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("HX-Request", "true")

	if !IsHTMX(req) {
		t.Error("IsHTMX = false, want true")
	}
}

func TestSetTrigger(t *testing.T) {
	rr := httptest.NewRecorder()
	SetTrigger(rr, "contactSaved")
	if got := rr.Header().Get("HX-Trigger"); got != "contactSaved" {
		t.Errorf("HX-Trigger = %q, want contactSaved", got)
	}
}

func TestSetToast(t *testing.T) {
	rr := httptest.NewRecorder()
	SetToast(rr, "Saved!")
	got := rr.Header().Get("HX-Trigger")
	if got == "" {
		t.Fatal("HX-Trigger empty")
	}
	if got != `{"caisToast":"Saved!"}` {
		t.Errorf("HX-Trigger = %q, want JSON caisToast payload", got)
	}
}

func TestSetFocus(t *testing.T) {
	rr := httptest.NewRecorder()
	SetFocus(rr, "#email")
	got := rr.Header().Get("HX-Trigger")
	if got == "" {
		t.Fatal("HX-Trigger empty")
	}
	if got != `{"caisFocus":"#email"}` {
		t.Errorf("HX-Trigger = %q, want caisFocus payload", got)
	}
}

func TestSetRetarget(t *testing.T) {
	rr := httptest.NewRecorder()
	SetRetarget(rr, "#form-errors")
	if got := rr.Header().Get("HX-Retarget"); got != "#form-errors" {
		t.Errorf("HX-Retarget = %q, want #form-errors", got)
	}
}

func TestIsHTMX_False(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	if IsHTMX(req) {
		t.Error("IsHTMX = true, want false")
	}
}
