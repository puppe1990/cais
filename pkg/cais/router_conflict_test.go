package cais

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRegister_ConflictingPatterns_PanicsWithHint(t *testing.T) {
	r := NewRouter()
	r.Post("/webhooks/kiwify/{token}", func(http.ResponseWriter, *http.Request) {})

	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic on conflicting ServeMux patterns")
		}
		msg := fmt.Sprint(rec)
		if !strings.Contains(msg, "cais:") {
			t.Errorf("panic should include cais: hint, got: %v", rec)
		}
		if !strings.Contains(msg, "DELETE") && !strings.Contains(msg, "distinct prefix") {
			t.Errorf("panic should suggest safer route shapes, got: %v", rec)
		}
		if !strings.Contains(msg, "/webhooks/{id}/delete") {
			t.Errorf("panic should mention the new pattern, got: %v", rec)
		}
	}()

	r.Post("/webhooks/{id}/delete", func(http.ResponseWriter, *http.Request) {})
}

func TestRegister_NonConflictingWildcardOK(t *testing.T) {
	r := NewRouter()
	// Literal sibling is more specific — ServeMux allows this.
	r.Get("/users/{id}", func(http.ResponseWriter, *http.Request) {})
	r.Get("/users/me", func(http.ResponseWriter, *http.Request) {})
}

func TestPatternsConflict_servemuxExamples(t *testing.T) {
	cases := []struct {
		a, b    string
		want    bool
		comment string
	}{
		{"/webhooks/{id}/delete", "/webhooks/kiwify/{token}", true, "classic conflict"},
		{"/users/{id}", "/users/me", false, "literal more specific"},
		{"/users/{id}", "/posts/{id}", false, "different first segment"},
		{"/webhooks/{id}", "/webhooks/{id}/delete", false, "different depth"},
		{"/a/{x}/b", "/a/{y}/b", true, "identical shape wildcards"},
		{"/a/b", "/a/b", true, "duplicate"},
	}
	for _, tc := range cases {
		got := patternsConflict(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("patternsConflict(%q, %q) = %v, want %v (%s)", tc.a, tc.b, got, tc.want, tc.comment)
		}
		// commutative
		if patternsConflict(tc.b, tc.a) != tc.want {
			t.Errorf("patternsConflict not commutative for %q %q", tc.a, tc.b)
		}
	}
}
