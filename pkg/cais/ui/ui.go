package ui

import (
	"fmt"
	"html/template"
	"strings"
)

const (
	navTabBase = "px-3 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 flex-shrink-0"
	navTabOn   = "bg-slate-900 text-white shadow-2xs"
	navTabOff  = "text-slate-600 hover:text-slate-900 hover:bg-slate-100"
)

// NavTabData describes a horizontal tab link in app layouts.
type NavTabData struct {
	Href   string
	Label  string
	Icon   string // optional; see iconPaths
	Active bool
}

// Funcs returns template helpers for layout chrome (tabs, icons).
func Funcs() template.FuncMap {
	return template.FuncMap{
		"navTabClass": NavTabClass,
		"icon":        Icon,
		"navTab":      NavTab,
		"makeNavTab":  MakeNavTab,
	}
}

// NavTabClass returns Tailwind classes for a tab pill.
func NavTabClass(active bool) string {
	if active {
		return navTabBase + " " + navTabOn
	}
	return navTabBase + " " + navTabOff
}

// MakeNavTab builds NavTabData for templates.
func MakeNavTab(href, label, icon string, active bool) NavTabData {
	return NavTabData{Href: href, Label: label, Icon: icon, Active: active}
}

// NavTab renders a complete tab anchor with optional icon and hx-boost shell navigation.
func NavTab(tab NavTabData) template.HTML {
	var b strings.Builder
	b.WriteString(`<a href="`)
	b.WriteString(template.HTMLEscapeString(tab.Href))
	b.WriteString(`" data-cais-nav="`)
	b.WriteString(template.HTMLEscapeString(tab.Href))
	b.WriteString(`" hx-boost="true" hx-target="#cais-main" hx-select="#cais-main" hx-swap="innerHTML swap:150ms" data-cais-view-transition class="`)
	b.WriteString(NavTabClass(tab.Active))
	b.WriteString(`">`)
	if tab.Icon != "" {
		b.WriteString(string(Icon(tab.Icon, "w-3.5 h-3.5")))
	}
	b.WriteString(template.HTMLEscapeString(tab.Label))
	b.WriteString(`</a>`)
	return template.HTML(b.String())
}

// Icon returns an inline SVG for a known name, or empty when unknown.
func Icon(name, class string) template.HTML {
	paths, ok := iconPaths[name]
	if !ok {
		return ""
	}
	if class == "" {
		class = "w-4 h-4"
	}
	return template.HTML(fmt.Sprintf(
		`<svg class="%s" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">%s</svg>`,
		template.HTMLEscapeString(class),
		paths,
	))
}
