package cli

import (
	"strings"
	"testing"
)

func TestLayoutTemplates_containNavMarker(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		if !strings.Contains(tpl, "<!-- cais:nav -->") {
			t.Errorf("%s layout missing <!-- cais:nav --> marker", name)
		}
	}
}

func TestLayoutTemplates_fullHasDefaultNavLinks(t *testing.T) {
	for _, link := range []string{`makeNavTab "/contact"`, `makeNavTab "/dashboard"`, `makeNavTab "/"`} {
		if !strings.Contains(tplLayout, link) {
			t.Errorf("full layout missing %s", link)
		}
	}
	if !strings.Contains(tplLayout, "navTab") || !strings.Contains(tplLayout, "cais-toast-host") {
		t.Error("full layout should use navTab helpers and cais-toast-host")
	}
}

func TestLayoutTemplates_minimalAndBlankMatch(t *testing.T) {
	if tplLayoutMinimal != tplLayoutBlank {
		t.Error("minimal and blank base layouts should be identical")
	}
}

func TestLayoutTemplates_useSharedHead(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		if !strings.Contains(tpl, "htmx.min.js") || !strings.Contains(tpl, `define "base"`) {
			t.Errorf("%s layout missing shared head/shell fragments", name)
		}
		if !strings.Contains(tpl, "idiomorph-ext.min.js") || !strings.Contains(tpl, `hx-ext="morph"`) {
			t.Errorf("%s layout missing idiomorph morph extension", name)
		}
	}
}

func TestLayoutTemplates_hasBoostShell(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		if !strings.Contains(tpl, `id="cais-nav"`) {
			t.Errorf("%s layout missing cais-nav id", name)
		}
		if !strings.Contains(tpl, `id="cais-main"`) || !strings.Contains(tpl, `data-cais-view-transition`) {
			t.Errorf("%s layout missing cais-main shell with view transition", name)
		}
	}
}

func TestLayoutTemplates_navTabsHaveIcons(t *testing.T) {
	for _, icon := range []string{`"home"`, `"message"`, `"chart"`} {
		if !strings.Contains(tplLayout, icon) {
			t.Errorf("full layout nav should include icon %s", icon)
		}
	}
}

func TestLayoutTemplates_contactFormUsesHxFormHelper(t *testing.T) {
	if !strings.Contains(tplPageContact, `hxForm "/contact"`) {
		t.Error("contact form should use hxForm helper")
	}
}

func TestLayoutTemplates_dashboardUsesIconHelper(t *testing.T) {
	for _, icon := range []string{`icon "users"`, `icon "shield"`} {
		if !strings.Contains(tplPageDashboard, icon) {
			t.Errorf("dashboard page should use %s helper", icon)
		}
	}
}

func TestLayoutTemplates_supermarketDesignTokens(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		for _, token := range []string{
			"font-display",
			"shadow-2xs",
			"no-scrollbar",
			"max-w-7xl",
			"sticky top-0",
		} {
			if !strings.Contains(tpl, token) {
				t.Errorf("%s layout missing design token %q", name, token)
			}
		}
	}
}
