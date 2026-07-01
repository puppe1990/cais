package meta

import (
	"strings"
	"testing"
)

func TestAbsoluteURL(t *testing.T) {
	tests := []struct {
		base, path, want string
	}{
		{"", "/static/og.png", "/static/og.png"},
		{"https://pulsefit.gestaobem.com", "/static/og.png", "https://pulsefit.gestaobem.com/static/og.png"},
		{"https://pulsefit.gestaobem.com/", "static/og.png", "https://pulsefit.gestaobem.com/static/og.png"},
		{"https://pulsefit.gestaobem.com", "", "https://pulsefit.gestaobem.com/"},
	}
	for _, tc := range tests {
		if got := AbsoluteURL(tc.base, tc.path); got != tc.want {
			t.Errorf("AbsoluteURL(%q, %q) = %q, want %q", tc.base, tc.path, got, tc.want)
		}
	}
}

func TestDefaultPreview(t *testing.T) {
	p := DefaultPreview("PulseFit")
	if p.Title != "PulseFit" {
		t.Errorf("Title = %q", p.Title)
	}
	if p.Image != DefaultImagePath {
		t.Errorf("Image = %q", p.Image)
	}
	if p.Locale != "pt_BR" {
		t.Errorf("Locale = %q", p.Locale)
	}
}

func TestPreviewHTML(t *testing.T) {
	html := PreviewHTML(Preview{
		Title:       "PulseFit",
		Description: "Log workouts with rhythm.",
		SiteName:    "PulseFit",
		SiteURL:     "https://pulsefit.gestaobem.com",
		Path:        "/login",
		Image:       DefaultImagePath,
	})

	for _, want := range []string{
		`property="og:type" content="website"`,
		`property="og:title" content="PulseFit"`,
		`property="og:description" content="Log workouts with rhythm."`,
		`property="og:image" content="https://pulsefit.gestaobem.com/static/og.png"`,
		`property="og:url" content="https://pulsefit.gestaobem.com/login"`,
		`name="twitter:card" content="summary_large_image"`,
		`name="twitter:image" content="https://pulsefit.gestaobem.com/static/og.png"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("PreviewHTML missing %q in:\n%s", want, html)
		}
	}
}

func TestSiteFrom(t *testing.T) {
	site := SiteFrom("My App", "https://example.com/")
	if site.AppName != "My App" {
		t.Errorf("AppName = %q", site.AppName)
	}
	if site.AppURL != "https://example.com" {
		t.Errorf("AppURL = %q", site.AppURL)
	}
}

func TestTemplateFuncs_absURL(t *testing.T) {
	fn, ok := TemplateFuncs()["absURL"].(func(string, string) string)
	if !ok {
		t.Fatal("absURL func missing")
	}
	if got := fn("https://example.com", "/static/og.png"); got != "https://example.com/static/og.png" {
		t.Errorf("absURL = %q", got)
	}
}
