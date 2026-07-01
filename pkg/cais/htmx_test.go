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

func TestIsHTMX_False(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	if IsHTMX(req) {
		t.Error("IsHTMX = true, want false")
	}
}