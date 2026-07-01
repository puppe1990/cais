package testutil

import (
	"net/http"
	"testing"
)

func TestNewRequest_withPathValue(t *testing.T) {
	req := NewRequest(http.MethodGet, "/items/42", PathValue("id", "42"))
	if req.PathValue("id") != "42" {
		t.Errorf("id = %q, want 42", req.PathValue("id"))
	}
}

func TestNewRenderer_parsesTemplates(t *testing.T) {
	renderer := NewRenderer(t)
	if renderer == nil {
		t.Fatal("renderer is nil")
	}
}
