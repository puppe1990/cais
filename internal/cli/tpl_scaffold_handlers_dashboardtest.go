// Inertia handler scaffold templates (ported from Cais demo).
package cli

const tplDashboardTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
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
`
