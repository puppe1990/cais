package i18n

import (
	"net/http"
	"strings"
)

// CookieName matches cais_csrf / cais_flash so locale is a first-party app cookie.
const CookieName = "cais_locale"

// CatalogForRequest picks a catalog from the request then fallback.
// Order: ?lang= query, cais_locale cookie, fallback, DefaultLocale, DefaultCatalog.
// Tags are normalized (pt-BR → pt) and must be keys in catalogs (allow-list).
// Unknown tags skip to the next source. Accept-Language is ignored.
func CatalogForRequest(r *http.Request, catalogs map[string]*Catalog, fallback string) *Catalog {
	if c := catalogFromQuery(r, catalogs); c != nil {
		return c
	}
	if c := catalogFromCookie(r, catalogs); c != nil {
		return c
	}
	if c := catalogForTag(fallback, catalogs); c != nil {
		return c
	}
	if c := catalogForTag(DefaultLocale, catalogs); c != nil {
		return c
	}
	return DefaultCatalog()
}

func catalogFromQuery(r *http.Request, catalogs map[string]*Catalog) *Catalog {
	if r == nil || r.URL == nil {
		return nil
	}
	return catalogForTag(r.URL.Query().Get("lang"), catalogs)
}

func catalogFromCookie(r *http.Request, catalogs map[string]*Catalog) *Catalog {
	if r == nil {
		return nil
	}
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return nil
	}
	return catalogForTag(cookie.Value, catalogs)
}

func catalogForTag(tag string, catalogs map[string]*Catalog) *Catalog {
	if catalogs == nil {
		return nil
	}
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil
	}
	c, ok := catalogs[allowListKey(tag)]
	if !ok || c == nil {
		return nil
	}
	return c
}

// allowListKey is the catalogs-map key for a request tag.
// Known families use NormalizeLocale (pt-BR → pt). Unknown tags keep their
// primary subtag so they miss the allow-list instead of collapsing to en.
func allowListKey(tag string) string {
	normalized := NormalizeLocale(tag)
	primary := primarySubtag(tag)
	if normalized == DefaultLocale && primary != "" && !strings.HasPrefix(primary, "en") {
		return primary
	}
	return normalized
}

func primarySubtag(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	tag = strings.ReplaceAll(tag, "-", "_")
	if i := strings.IndexByte(tag, '_'); i >= 0 {
		tag = tag[:i]
	}
	return tag
}
