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
		`hx-swap="innerHTML swap:150ms transition:true"`,
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
		`hx-push-url="true"`,
		`hx-swap="innerHTML swap:150ms transition:true"`,
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

func TestHxPost(t *testing.T) {
	got := string(HxPost("/feed/1/confirm", "#scan-actions-1"))
	for _, want := range []string{
		`hx-post="/feed/1/confirm"`,
		`hx-target="#scan-actions-1"`,
		`hx-swap="outerHTML"`,
		`data-cais-optimistic="count"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HxPost missing %q, got %q", want, got)
		}
	}
}

func TestHxPostConfirm(t *testing.T) {
	got := string(HxPostConfirm("/feed/1/flag", "this", "Reportar este preço?"))
	for _, want := range []string{
		`hx-post="/feed/1/flag"`,
		`hx-target="this"`,
		`hx-confirm="Reportar este preço?"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HxPostConfirm missing %q, got %q", want, got)
		}
	}
}

func TestHxChatForm(t *testing.T) {
	got := string(HxChatForm("/chat/1/messages", "#chat-thinking"))
	for _, want := range []string{
		`hx-post="/chat/1/messages"`,
		`hx-target="#chat-history"`,
		`hx-swap="beforeend"`,
		`chat-thinking`,
		`data-cais-chat-form`,
		`caisFinalizeChatStream`,
		`hx-disabled-elt="button[type='submit']"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HxChatForm missing %q, got %q", want, got)
		}
	}
	for _, omit := range []string{
		`hx-on:keydown`,
		`hx-on::after-request`,
		`this.reset()`,
	} {
		if strings.Contains(got, omit) {
			t.Errorf("HxChatForm should not contain %q, got %q", omit, got)
		}
	}
}

func TestFuncs_registersHelpers(t *testing.T) {
	fns := Funcs()
	for _, name := range []string{"hxForm", "hxChatForm", "hxDelete", "hxBoostLink", "hxPaginate", "hxMorphOuter", "hxPost", "hxPostConfirm"} {
		if fns[name] == nil {
			t.Errorf("Funcs() missing %q", name)
		}
	}
}
