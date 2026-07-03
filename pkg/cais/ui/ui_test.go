package ui

import (
	"strings"
	"testing"
)

func TestNavTabClass_active(t *testing.T) {
	got := NavTabClass(true)
	for _, want := range []string{"bg-slate-900", "text-white", "shadow-2xs"} {
		if !strings.Contains(got, want) {
			t.Errorf("NavTabClass(true) missing %q, got %q", want, got)
		}
	}
}

func TestNavTabClass_inactive(t *testing.T) {
	got := NavTabClass(false)
	for _, want := range []string{"text-slate-600", "hover:bg-slate-100"} {
		if !strings.Contains(got, want) {
			t.Errorf("NavTabClass(false) missing %q, got %q", want, got)
		}
	}
	if strings.Contains(got, "bg-slate-900") {
		t.Error("inactive tab should not include active background")
	}
}

func TestIcon_knownName(t *testing.T) {
	got := string(Icon("camera", "w-4 h-4"))
	if !strings.Contains(got, `<svg`) || !strings.Contains(got, `w-4 h-4`) {
		t.Fatalf("Icon(camera) = %q", got)
	}
	if !strings.Contains(got, `aria-hidden="true"`) {
		t.Error("icon should be aria-hidden")
	}
}

func TestIcon_unknownName(t *testing.T) {
	got := Icon("not-a-real-icon", "")
	if got != "" {
		t.Errorf("unknown icon should return empty, got %q", got)
	}
}

func TestNavTab_rendersLink(t *testing.T) {
	got := string(NavTab(NavTabData{Href: "/feed", Label: "Feed", Icon: "message", Active: true}))
	for _, want := range []string{`href="/feed"`, "Feed", "bg-slate-900", `<svg`} {
		if !strings.Contains(got, want) {
			t.Errorf("NavTab missing %q, got %q", want, got)
		}
	}
}

func TestNavTab_hxBoostShell(t *testing.T) {
	got := string(NavTab(NavTabData{Href: "/contact", Label: "Contact", Icon: "message", Active: false}))
	for _, want := range []string{
		`hx-boost="true"`,
		`hx-target="#cais-main"`,
		`hx-select="#cais-main"`,
		`data-cais-nav="/contact"`,
		`data-cais-view-transition`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("NavTab missing %q, got %q", want, got)
		}
	}
}

func TestMakeNavTab(t *testing.T) {
	tab := MakeNavTab("/map", "Mapa", "map", true)
	if tab.Href != "/map" || tab.Label != "Mapa" || tab.Icon != "map" || !tab.Active {
		t.Errorf("MakeNavTab: %+v", tab)
	}
}

func TestIcon_navAndDashboardIcons(t *testing.T) {
	for _, name := range []string{"message", "chart", "users", "shield"} {
		got := string(Icon(name, "w-5 h-5"))
		if !strings.Contains(got, `<svg`) {
			t.Errorf("Icon(%q) should render svg, got %q", name, got)
		}
	}
}

func TestFuncs_registersHelpers(t *testing.T) {
	fns := Funcs()
	for _, name := range []string{"navTabClass", "icon", "navTab", "makeNavTab"} {
		if fns[name] == nil {
			t.Errorf("Funcs() missing %q", name)
		}
	}
}
