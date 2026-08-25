package i18n

import (
	"context"
	"net/http"
	"strings"
)

type catalogCtxKey struct{}

const (
	// CookieName matches cais_csrf / cais_flash so locale is a first-party app cookie.
	CookieName = "cais_locale"
	// CookieMaxAge is one year so a switcher choice survives visits without
	// Accept-Language negotiation (out of scope for v1).
	CookieMaxAge = 86400 * 365
)

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

// LocaleMiddleware resolves the catalog for each request and stores it on
// context. It is read-only: ?lang= is not written to cais_locale here; apps
// persist a language switch via SetCookie on a dedicated /locale route.
func LocaleMiddleware(catalogs map[string]*Catalog, fallback string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cat := CatalogForRequest(r, catalogs, fallback)
			ctx := context.WithValue(r.Context(), catalogCtxKey{}, cat)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CatalogFromRequest returns the catalog stored by LocaleMiddleware, or nil.
func CatalogFromRequest(r *http.Request) *Catalog {
	if r == nil {
		return nil
	}
	c, _ := r.Context().Value(catalogCtxKey{}).(*Catalog)
	return c
}

// SetCookie writes cais_locale for app language-switcher handlers (GET /locale?lang=).
func SetCookie(w http.ResponseWriter, locale string, secure bool) {
	if strings.TrimSpace(locale) == "" {
		return
	}
	tag := NormalizeLocale(locale)
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    tag,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   secure,
	})
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
