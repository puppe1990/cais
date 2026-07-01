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
