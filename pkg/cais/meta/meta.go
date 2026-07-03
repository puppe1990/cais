package meta

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais/csrf"
	"github.com/puppe1990/cais/pkg/cais/flash"
)

const DefaultImagePath = "/static/og.png"

// Site carries app-level values available in layout templates.
type Site struct {
	AppName    string
	AppURL     string
	CSRFToken  string
	Flash      *flash.Message
	ActiveNav  string // optional tab key for layout navTab helpers (e.g. "home", "contact")
	UserLevel  int    // optional gamification chrome (0 = layout may show defaults)
	UserPoints int
	UserRank   int
	LoggedIn   bool // optional session flag for layout auth chrome
}

// Preview describes Open Graph / Twitter card metadata for a page.
type Preview struct {
	Title       string
	Description string
	SiteName    string
	SiteURL     string
	Path        string
	Image       string
	Locale      string
	Type        string
}

func DefaultPreview(name string) Preview {
	return Preview{
		Title:       name,
		Description: name + " — powered by Cais",
		SiteName:    name,
		Image:       DefaultImagePath,
		Locale:      "en_US",
		Type:        "website",
	}
}

func SiteFrom(appName, appURL string) Site {
	return Site{
		AppName: appName,
		AppURL:  strings.TrimRight(strings.TrimSpace(appURL), "/"),
	}
}

// WithCSRF returns site with the per-request CSRF token for layout templates.
func WithCSRF(site Site, r *http.Request) Site {
	site.CSRFToken = csrf.TokenFromRequest(r)
	return site
}

// ForRequest returns site with per-request CSRF and flash values for layout templates.
func ForRequest(site Site, r *http.Request) Site {
	site = WithCSRF(site, r)
	if msg, ok := flash.MessageFromRequest(r); ok {
		site.Flash = &msg
	}
	return site
}

// AbsoluteURL joins a site base URL with a path. When base is empty, returns path.
func AbsoluteURL(base, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return path
	}
	return base + path
}

// PreviewHTML returns Open Graph and Twitter meta tags for <head>.
func PreviewHTML(p Preview) string {
	if p.Type == "" {
		p.Type = "website"
	}
	if p.Locale == "" {
		p.Locale = "en_US"
	}
	if p.Image == "" {
		p.Image = DefaultImagePath
	}

	image := AbsoluteURL(p.SiteURL, p.Image)
	title := attr(p.Title)
	description := attr(p.Description)
	siteName := attr(p.SiteName)

	var b strings.Builder
	b.WriteString(`<meta name="description" content="`)
	b.WriteString(description)
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta property="og:type" content="`)
	b.WriteString(attr(p.Type))
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta property="og:site_name" content="`)
	b.WriteString(siteName)
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta property="og:title" content="`)
	b.WriteString(title)
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta property="og:description" content="`)
	b.WriteString(description)
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta property="og:image" content="`)
	b.WriteString(attr(image))
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta property="og:locale" content="`)
	b.WriteString(attr(p.Locale))
	b.WriteString(`" />` + "\n")
	if p.SiteURL != "" && p.Path != "" {
		b.WriteString(`    <meta property="og:url" content="`)
		b.WriteString(attr(AbsoluteURL(p.SiteURL, p.Path)))
		b.WriteString(`" />` + "\n")
	}
	b.WriteString(`    <meta name="twitter:card" content="summary_large_image" />` + "\n")
	b.WriteString(`    <meta name="twitter:title" content="`)
	b.WriteString(title)
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta name="twitter:description" content="`)
	b.WriteString(description)
	b.WriteString(`" />` + "\n")
	b.WriteString(`    <meta name="twitter:image" content="`)
	b.WriteString(attr(image))
	b.WriteString(`" />`)
	return b.String()
}

// TemplateFuncs returns helpers for layout templates.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"absURL": AbsoluteURL,
	}
}

func attr(v string) string {
	return template.HTMLEscapeString(v)
}

// WithPage returns a copy of preview with page-specific fields.
func WithPage(base Preview, title, description, path string) Preview {
	if title != "" {
		base.Title = title
	}
	if description != "" {
		base.Description = description
	}
	base.Path = path
	return base
}

// DefaultDescription returns the standard app description.
func DefaultDescription(name string) string {
	return fmt.Sprintf("%s — powered by Cais", name)
}
