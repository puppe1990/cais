// Inertia handler scaffold templates (ported from Cais demo).
package cli

const tplDashboardTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/flash"
)

func TestDashboardHandler_InertiaComponent(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{}, setupTestInertia(t))

	req := inertiaRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	assertInertiaComponent(t, rr, "Dashboard")
}

func TestDashboardHandler_includesFlashProp(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{}, setupTestInertia(t))

	req := inertiaRequest(http.MethodGet, "/dashboard", nil)
	req = flash.WithMessage(req, flash.Message{Kind: "notice", Message: "Welcome back!"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	flashProp, ok := assertInertiaProp(t, rr, "flash").(map[string]any)
	if !ok || flashProp["notice"] != "Welcome back!" {
		t.Errorf("props.flash missing notice: %v", flashProp)
	}
}
`
