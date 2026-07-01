package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func testRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	r, err := cais.NewRendererFromDir("../testdata/templates")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestSeeOther(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/old", nil)
	SeeOther(rr, req, "/new")
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/new" {
		t.Errorf("Location = %q, want /new", loc)
	}
}

func TestRenderOrError_rendersPage(t *testing.T) {
	renderer := testRenderer(t)
	rr := httptest.NewRecorder()
	RenderOrError(rr, renderer, "base", "home", map[string]string{"Name": "Test"})
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Test") {
		t.Errorf("body missing content: %s", rr.Body.String())
	}
}

func TestRenderPartial_rendersFragment(t *testing.T) {
	renderer := testRenderer(t)
	rr := httptest.NewRecorder()
	if err := RenderPartial(rr, renderer, "greeting", map[string]string{"Name": "Ada"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rr.Body.String(), "Ada") {
		t.Errorf("body = %q", rr.Body.String())
	}
}
