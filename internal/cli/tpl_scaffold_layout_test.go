package cli

import (
	"os"
	"path/filepath"
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
	if !strings.Contains(tplLayout, `template "nav_links"`) {
		t.Error("full layout should render nav_links partial")
	}
	for _, link := range []string{`href="/contact"`, `href="/dashboard"`, `hx-boost`} {
		if !strings.Contains(tplPartialNavLinks, link) {
			t.Errorf("nav_links partial missing %s", link)
		}
	}
	if !strings.Contains(tplLayout, "cais-toast-host") {
		t.Error("full layout missing cais-toast-host")
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
	for _, icon := range []string{`icon_home_nav`, `icon_message_nav`, `icon_chart_nav`} {
		if !strings.Contains(tplPartialNavLinks, icon) {
			t.Errorf("nav partial should include %s", icon)
		}
	}
}

func TestScaffoldPartials_iconsRenderNonEmpty(t *testing.T) {
	dir := t.TempDir()
	data := scaffoldData{AppName: "demo", ModulePath: "github.com/acme/demo"}
	for path, tpl := range map[string]string{
		"web/templates/partials/icons.html":     tplPartialIcons,
		"web/templates/partials/nav_links.html": tplPartialNavLinks,
	} {
		if err := writeTemplate(filepath.Join(dir, path), tpl, data); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		body, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatal(err)
		}
		if len(body) == 0 {
			t.Fatalf("%s rendered empty", path)
		}
		want := `define "icon_sparkles_md"`
		if path == "web/templates/partials/nav_links.html" {
			want = `define "nav_links"`
		}
		if !strings.Contains(string(body), want) {
			t.Fatalf("%s missing %s", path, want)
		}
	}
}

func TestLayoutTemplates_contactFormUsesHxFormHelper(t *testing.T) {
	if !strings.Contains(tplPageContact, `hxForm "/contact"`) {
		t.Error("contact form should use hxForm helper")
	}
}

func TestLayoutTemplates_dashboardUsesIconPartials(t *testing.T) {
	for _, icon := range []string{`icon_users_md`, `icon_shield_md`} {
		if !strings.Contains(tplPageDashboard, icon) {
			t.Errorf("dashboard page should use %s partial", icon)
		}
	}
}

func TestLayoutTemplates_shellDesignTokens(t *testing.T) {
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
