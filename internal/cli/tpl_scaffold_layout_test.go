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
	for _, link := range []string{`href="/contact"`, `href="/dashboard"`, `href="/"`} {
		if !strings.Contains(tplLayout, link) {
			t.Errorf("full layout missing %s", link)
		}
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
	}
}
