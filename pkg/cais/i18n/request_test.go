package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeLocale_public(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"pt-BR", "pt"},
		{"es-MX", "es"},
		{"zh-CN", "zh"},
		{"", "en"},
	}
	for _, tc := range tests {
		if got := NormalizeLocale(tc.in); got != tc.want {
			t.Errorf("NormalizeLocale(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCatalogForRequest_queryLangSelectsCatalog(t *testing.T) {
	req := newLocaleRequest(t, "/?lang=pt")
	got := CatalogForRequest(req, requestCatalogs(), "en")
	assertLocale(t, got, "pt")
}

func TestCatalogForRequest_normalizesQueryTag(t *testing.T) {
	req := newLocaleRequest(t, "/?lang=pt-BR")
	got := CatalogForRequest(req, requestCatalogs(), "en")
	assertLocale(t, got, "pt")
}

func TestCatalogForRequest_cookieSelectsCatalog(t *testing.T) {
	req := newLocaleRequest(t, "/")
	req.AddCookie(&http.Cookie{Name: "cais_locale", Value: "pt"})
	got := CatalogForRequest(req, requestCatalogs(), "en")
	assertLocale(t, got, "pt")
}

func TestCatalogForRequest_queryWinsOverCookie(t *testing.T) {
	req := newLocaleRequest(t, "/?lang=en")
	req.AddCookie(&http.Cookie{Name: "cais_locale", Value: "pt"})
	got := CatalogForRequest(req, requestCatalogs(), "pt")
	assertLocale(t, got, "en")
}

func TestCatalogForRequest_unknownLangUsesFallback(t *testing.T) {
	req := newLocaleRequest(t, "/?lang=xx")
	got := CatalogForRequest(req, requestCatalogs(), "pt")
	assertLocale(t, got, "pt")
}

func TestCatalogForRequest_acceptLanguageIgnored(t *testing.T) {
	req := newLocaleRequest(t, "/")
	req.Header.Set("Accept-Language", "pt")
	got := CatalogForRequest(req, requestCatalogs(), "en")
	assertLocale(t, got, "en")
}

func TestCatalogForRequest_emptyMap_returnsDefaultCatalog(t *testing.T) {
	req := newLocaleRequest(t, "/?lang=pt")
	got := CatalogForRequest(req, nil, "pt")
	if got == nil {
		t.Fatal("nil catalogs returned nil")
	}
	assertLocale(t, got, DefaultLocale)

	got = CatalogForRequest(req, map[string]*Catalog{}, "pt")
	if got == nil {
		t.Fatal("empty catalogs returned nil")
	}
	assertLocale(t, got, DefaultLocale)
}

func TestLocaleMiddleware_storesCatalogOnContext(t *testing.T) {
	var got string
	h := LocaleMiddleware(requestCatalogs(), "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := CatalogFromRequest(r)
		if c == nil {
			t.Fatal("CatalogFromRequest returned nil")
		}
		got = c.Locale()
	}))
	req := newLocaleRequest(t, "/?lang=pt")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "pt" {
		t.Errorf("CatalogFromRequest locale = %q, want pt", got)
	}
}

func TestLocaleMiddleware_usesCookieWhenNoQuery(t *testing.T) {
	var got string
	h := LocaleMiddleware(requestCatalogs(), "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = CatalogFromRequest(r).Locale()
	}))
	req := newLocaleRequest(t, "/")
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "es"})
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != "es" {
		t.Errorf("CatalogFromRequest locale = %q, want es", got)
	}
}

func TestLocaleMiddleware_doesNotSetCookie(t *testing.T) {
	h := LocaleMiddleware(requestCatalogs(), "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, newLocaleRequest(t, "/?lang=pt"))
	for _, c := range rr.Result().Cookies() {
		if c.Name == CookieName {
			t.Fatalf("middleware set %s=%q; persist ?lang= from the app /locale route", c.Name, c.Value)
		}
	}
}

func TestSetCookie_writesCaisLocale(t *testing.T) {
	rr := httptest.NewRecorder()
	SetCookie(rr, "pt-BR", false)
	c := localeCookie(t, rr)
	if c.Name != CookieName || CookieName != "cais_locale" {
		t.Errorf("Name = %q, want cais_locale", c.Name)
	}
	if c.Value != "pt" {
		t.Errorf("Value = %q, want pt", c.Value)
	}
	if c.Path != "/" {
		t.Errorf("Path = %q, want /", c.Path)
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", c.SameSite)
	}
	if !c.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}
	if c.Secure {
		t.Error("Secure = true, want false")
	}
	if c.MaxAge != 86400*365 {
		t.Errorf("MaxAge = %d, want 365 days", c.MaxAge)
	}
}

func TestSetCookie_secureFlag(t *testing.T) {
	rr := httptest.NewRecorder()
	SetCookie(rr, "es", true)
	c := localeCookie(t, rr)
	if c.Value != "es" {
		t.Errorf("Value = %q, want es", c.Value)
	}
	if !c.Secure {
		t.Error("Secure = false, want true")
	}
}

func TestSetCookie_emptyLocaleSkipped(t *testing.T) {
	rr := httptest.NewRecorder()
	SetCookie(rr, "", false)
	SetCookie(rr, "   ", false)
	for _, c := range rr.Result().Cookies() {
		if c.Name == CookieName {
			t.Fatalf("set cookie %s=%q for empty locale", c.Name, c.Value)
		}
	}
}

func TestCatalogFromRequest_nilWithoutMiddleware(t *testing.T) {
	req := newLocaleRequest(t, "/?lang=pt")
	if c := CatalogFromRequest(req); c != nil {
		t.Errorf("CatalogFromRequest = %v, want nil", c)
	}
}

func TestCatalogForRequest_esAndZhFromCustomMaps(t *testing.T) {
	catalogs := requestCatalogs()
	esReq := newLocaleRequest(t, "/?lang=es-MX")
	assertLocale(t, CatalogForRequest(esReq, catalogs, "en"), "es")
	zhReq := newLocaleRequest(t, "/?lang=zh-CN")
	assertLocale(t, CatalogForRequest(zhReq, catalogs, "en"), "zh")
}

func newLocaleRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func requestCatalogs() map[string]*Catalog {
	msgs := map[string]map[string]string{
		"en": {"hi": "Hello"},
		"pt": {"hi": "Olá"},
		"es": {"hi": "Hola"},
		"zh": {"hi": "你好"},
	}
	out := make(map[string]*Catalog, len(msgs))
	for tag := range msgs {
		out[tag] = NewCatalogFrom(tag, msgs)
	}
	return out
}

func localeCookie(t *testing.T, rr *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rr.Result().Cookies() {
		if c.Name == CookieName {
			return c
		}
	}
	t.Fatal("cais_locale cookie missing")
	return nil
}

func assertLocale(t *testing.T, c *Catalog, want string) {
	t.Helper()
	if c == nil {
		t.Fatalf("catalog is nil, want locale %q", want)
	}
	if got := c.Locale(); got != want {
		t.Errorf("Locale() = %q, want %q", got, want)
	}
}
