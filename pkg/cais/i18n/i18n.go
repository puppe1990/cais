package i18n

import (
	"fmt"
	"html/template"
	"strings"
)

const DefaultLocale = "en"

var locales = map[string]map[string]string{
	"en": enMessages,
	"pt": ptMessages,
}

// Catalog holds translated strings for a locale.
type Catalog struct {
	locale   string
	messages map[string]string
}

// DefaultCatalog returns the English catalog.
func DefaultCatalog() *Catalog {
	return NewCatalog(DefaultLocale)
}

// NewCatalog builds a catalog for the given locale tag.
func NewCatalog(locale string) *Catalog {
	return NewCatalogFrom(locale, locales)
}

// NewCatalogFrom builds a catalog from custom locale maps (for app-level translations).
func NewCatalogFrom(locale string, locales map[string]map[string]string) *Catalog {
	tag := normalizeLocale(locale)
	msgs, ok := locales[tag]
	if !ok {
		tag = DefaultLocale
		msgs, ok = locales[tag]
		if !ok {
			return &Catalog{locale: DefaultLocale, messages: map[string]string{}}
		}
	}
	copied := make(map[string]string, len(msgs))
	for k, v := range msgs {
		copied[k] = v
	}
	return &Catalog{locale: tag, messages: copied}
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	locale = strings.ReplaceAll(locale, "-", "_")
	switch {
	case locale == "" || strings.HasPrefix(locale, "en"):
		return "en"
	case strings.HasPrefix(locale, "pt"):
		return "pt"
	default:
		return DefaultLocale
	}
}

// T returns the translation for key, optionally formatting with args.
func (c *Catalog) T(key string, args ...any) string {
	msg, ok := c.messages[key]
	if !ok {
		return key
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

// Locale returns the normalized locale tag (en, pt).
func (c *Catalog) Locale() string {
	return c.locale
}

// HTMLLang returns the BCP 47 language tag for <html lang>.
func (c *Catalog) HTMLLang() string {
	if c.locale == "pt" {
		return "pt-BR"
	}
	return "en"
}

// OGLocale returns the Open Graph locale value.
func (c *Catalog) OGLocale() string {
	if c.locale == "pt" {
		return "pt_BR"
	}
	return "en_US"
}

// Funcs returns template helpers: t, htmlLang, ogLocale.
func (c *Catalog) Funcs() template.FuncMap {
	return template.FuncMap{
		"t": func(key string, args ...any) string {
			return c.T(key, args...)
		},
		"htmlLang": func() string { return c.HTMLLang() },
		"ogLocale": func() string { return c.OGLocale() },
	}
}

// MergeFuncs combines i18n funcs with additional template funcs.
func MergeFuncs(c *Catalog, extra template.FuncMap) template.FuncMap {
	out := template.FuncMap{}
	for k, v := range extra {
		out[k] = v
	}
	for k, v := range c.Funcs() {
		out[k] = v
	}
	return out
}
