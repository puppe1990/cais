package htmxattrs

import (
	"strings"
	"testing"
)

func TestHxForm(t *testing.T) {
	got := string(HxForm("/contact", "#form-errors", "#contact-spinner"))
	for _, want := range []string{
		`hx-post="/contact"`,
		`hx-target="#form-errors"`,
		`hx-swap="innerHTML swap:150ms"`,
		`data-cais-view-transition`,
		`hx-indicator="#contact-spinner"`,
		`hx-disabled-elt="button[type='submit']"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HxForm missing %q, got %q", want, got)
		}
	}
}

func TestHxDelete(t *testing.T) {
	got := string(HxDelete("/admin/items/1", "Delete this item?"))
	for _, want := range []string{
		`hx-delete="/admin/items/1"`,
		`hx-target="closest tr"`,
		`hx-swap="outerHTML swap:150ms"`,
		`hx-confirm="Delete this item?"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HxDelete missing %q, got %q", want, got)
		}
	}
}

func TestHxBoostLink(t *testing.T) {
	got := string(HxBoostLink())
	for _, want := range []string{
		`hx-boost="true"`,
		`hx-target="#cais-main"`,
		`hx-select="#cais-main"`,
		`data-cais-view-transition`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HxBoostLink missing %q, got %q", want, got)
		}
	}
}

func TestHxPaginate(t *testing.T) {
	got := string(HxPaginate("/admin/items?page=2", "#admin-items"))
	for _, want := range []string{
		`hx-get="/admin/items?page=2"`,
		`hx-target="#admin-items"`,
		`hx-swap="morph:innerHTML"`,
		`hx-push-url="true"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HxPaginate missing %q, got %q", want, got)
		}
	}
}

func TestHxMorphOuter(t *testing.T) {
	got := string(HxMorphOuter())
	if got != `hx-swap="morph:outerHTML"` {
		t.Errorf("HxMorphOuter = %q", got)
	}
}

func TestFuncs_registersHelpers(t *testing.T) {
	fns := Funcs()
	for _, name := range []string{"hxForm", "hxDelete", "hxBoostLink", "hxPaginate", "hxMorphOuter"} {
		if fns[name] == nil {
			t.Errorf("Funcs() missing %q", name)
		}
	}
}
