package i18n

import (
	"net/http"
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

func assertLocale(t *testing.T, c *Catalog, want string) {
	t.Helper()
	if c == nil {
		t.Fatalf("catalog is nil, want locale %q", want)
	}
	if got := c.Locale(); got != want {
		t.Errorf("Locale() = %q, want %q", got, want)
	}
}
