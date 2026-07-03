package htmxattrs

import (
	"html/template"
	"strings"
)

// Funcs returns template helpers that emit common HTMX attribute bundles.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"hxForm":        HxForm,
		"hxDelete":      HxDelete,
		"hxBoostLink":   HxBoostLink,
		"hxPaginate":    HxPaginate,
		"hxMorphOuter":  HxMorphOuter,
		"hxPost":        HxPost,
		"hxPostConfirm": HxPostConfirm,
	}
}

// HxForm returns attributes for an HTMX form that swaps errors into a target.
func HxForm(postURL, target, indicator string) template.HTMLAttr {
	var b strings.Builder
	b.WriteString(`hx-post="`)
	b.WriteString(template.HTMLEscapeString(postURL))
	b.WriteString(`" hx-target="`)
	b.WriteString(template.HTMLEscapeString(target))
	b.WriteString(`" hx-swap="innerHTML swap:150ms transition:true" data-cais-view-transition hx-disabled-elt="button[type='submit']"`)
	if indicator != "" {
		b.WriteString(` hx-indicator="`)
		b.WriteString(template.HTMLEscapeString(indicator))
		b.WriteString(`"`)
	}
	return template.HTMLAttr(b.String())
}

// HxDelete returns attributes for inline row delete with confirmation.
func HxDelete(url, confirm string) template.HTMLAttr {
	var b strings.Builder
	b.WriteString(`hx-delete="`)
	b.WriteString(template.HTMLEscapeString(url))
	b.WriteString(`" hx-target="closest tr" hx-swap="outerHTML swap:150ms" hx-confirm="`)
	b.WriteString(template.HTMLEscapeString(confirm))
	b.WriteString(`"`)
	return template.HTMLAttr(b.String())
}

// HxBoostLink returns attributes for SPA-like navigation into #cais-main.
func HxBoostLink() template.HTMLAttr {
	return template.HTMLAttr(`hx-boost="true" hx-target="#cais-main" hx-select="#cais-main" hx-push-url="true" hx-swap="innerHTML swap:150ms transition:true" data-cais-view-transition`)
}

// HxPaginate returns attributes for HTMX pagination with morph swap and URL push.
func HxPaginate(getURL, target string) template.HTMLAttr {
	var b strings.Builder
	b.WriteString(`hx-get="`)
	b.WriteString(template.HTMLEscapeString(getURL))
	b.WriteString(`" hx-target="`)
	b.WriteString(template.HTMLEscapeString(target))
	b.WriteString(`" hx-swap="morph:innerHTML" hx-push-url="true"`)
	return template.HTMLAttr(b.String())
}

// HxMorphOuter returns morph swap for single-element updates (e.g. bool toggles).
func HxMorphOuter() template.HTMLAttr {
	return template.HTMLAttr(`hx-swap="morph:outerHTML"`)
}

// HxPost returns attributes for an optimistic count POST into a target element.
func HxPost(postURL, target string) template.HTMLAttr {
	var b strings.Builder
	b.WriteString(`hx-post="`)
	b.WriteString(template.HTMLEscapeString(postURL))
	b.WriteString(`" hx-target="`)
	b.WriteString(template.HTMLEscapeString(target))
	b.WriteString(`" hx-swap="outerHTML" data-cais-optimistic="count"`)
	return template.HTMLAttr(b.String())
}

// HxPostConfirm returns attributes for a confirmed POST that replaces the triggering element.
func HxPostConfirm(postURL, target, confirm string) template.HTMLAttr {
	var b strings.Builder
	b.WriteString(`hx-post="`)
	b.WriteString(template.HTMLEscapeString(postURL))
	b.WriteString(`" hx-target="`)
	b.WriteString(template.HTMLEscapeString(target))
	b.WriteString(`" hx-swap="outerHTML" hx-confirm="`)
	b.WriteString(template.HTMLEscapeString(confirm))
	b.WriteString(`"`)
	return template.HTMLAttr(b.String())
}
