package i18n

import (
	"html/template"
	"strings"
	"testing"
)

func TestDefaultLocale_isEn(t *testing.T) {
	if DefaultLocale != "en" {
		t.Errorf("DefaultLocale = %q, want en", DefaultLocale)
	}
}

func TestDefaultCatalog_usesEnglish(t *testing.T) {
	c := DefaultCatalog()
	if got := c.T("auth.invalid_credentials"); got != "Invalid email or password." {
		t.Errorf("T(auth.invalid_credentials) = %q", got)
	}
}

func TestNewCatalog_en(t *testing.T) {
	c := NewCatalog("en")
	if got := c.T("contact.name_required"); got != "Name is required." {
		t.Errorf("T(contact.name_required) = %q", got)
	}
	if c.HTMLLang() != "en" {
		t.Errorf("HTMLLang() = %q, want en", c.HTMLLang())
	}
	if c.OGLocale() != "en_US" {
		t.Errorf("OGLocale() = %q, want en_US", c.OGLocale())
	}
}

func TestNewCatalog_pt(t *testing.T) {
	c := NewCatalog("pt_BR")
	if got := c.T("auth.welcome"); got != "Bem-vindo!" {
		t.Errorf("T(auth.welcome) = %q", got)
	}
	if c.HTMLLang() != "pt-BR" {
		t.Errorf("HTMLLang() = %q, want pt-BR", c.HTMLLang())
	}
	if c.OGLocale() != "pt_BR" {
		t.Errorf("OGLocale() = %q, want pt_BR", c.OGLocale())
	}
}

func TestNewCatalog_emptyLocale_defaultsToEn(t *testing.T) {
	c := NewCatalog("")
	if got := c.T("auth.welcome"); got != "Welcome!" {
		t.Errorf("T(auth.welcome) = %q, want Welcome!", got)
	}
}

func TestCatalog_T_withArgs(t *testing.T) {
	c := NewCatalog("en")
	got := c.T("home.welcome", "Alice")
	if got != "Welcome, Alice!" {
		t.Errorf("T(home.welcome) = %q", got)
	}
}

func TestCatalog_T_missingKey_returnsKey(t *testing.T) {
	c := NewCatalog("en")
	if got := c.T("missing.key"); got != "missing.key" {
		t.Errorf("T(missing.key) = %q, want missing.key", got)
	}
}

func TestCatalog_Funcs_t(t *testing.T) {
	c := NewCatalog("en")
	fn, ok := c.Funcs()["t"].(func(string, ...any) string)
	if !ok {
		t.Fatal("t func missing or wrong type")
	}
	if got := fn("layout.footer"); got != "Running light on Lightsail" {
		t.Errorf("t(layout.footer) = %q", got)
	}
}

func TestCatalog_Funcs_htmlLangAndOgLocale(t *testing.T) {
	c := NewCatalog("en")
	htmlLang, ok := c.Funcs()["htmlLang"].(func() string)
	if !ok {
		t.Fatal("htmlLang func missing")
	}
	if htmlLang() != "en" {
		t.Errorf("htmlLang() = %q", htmlLang())
	}
	ogLocale, ok := c.Funcs()["ogLocale"].(func() string)
	if !ok {
		t.Fatal("ogLocale func missing")
	}
	if ogLocale() != "en_US" {
		t.Errorf("ogLocale() = %q", ogLocale())
	}
}

func TestMergeFuncs_includesMetaAndI18n(t *testing.T) {
	c := NewCatalog("en")
	funcs := MergeFuncs(c, template.FuncMap{"absURL": func(string, string) string { return "/x" }})
	if _, ok := funcs["absURL"]; !ok {
		t.Error("absURL missing")
	}
	if _, ok := funcs["t"]; !ok {
		t.Error("t missing")
	}
}

func TestNormalizeLocale(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "en"},
		{"en", "en"},
		{"en_US", "en"},
		{"en-US", "en"},
		{"pt", "pt"},
		{"pt_BR", "pt"},
		{"pt-BR", "pt"},
		{"fr", "en"},
	}
	for _, tc := range tests {
		if got := normalizeLocale(tc.in); got != tc.want {
			t.Errorf("normalizeLocale(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCatalog_enCoversAllKeys(t *testing.T) {
	pt := locales["pt"]
	en := locales["en"]
	for key := range pt {
		if _, ok := en[key]; !ok {
			t.Errorf("en missing key %q present in pt", key)
		}
	}
}

func TestNewCatalogFrom_customLocales(t *testing.T) {
	locales := map[string]map[string]string{
		"en": {"greeting": "Hello"},
		"pt": {"greeting": "Olá"},
	}
	c := NewCatalogFrom("pt-BR", locales)
	if got := c.T("greeting"); got != "Olá" {
		t.Errorf("T(greeting) = %q, want Olá", got)
	}
	if c.HTMLLang() != "pt-BR" {
		t.Errorf("HTMLLang() = %q, want pt-BR", c.HTMLLang())
	}
}

func TestCatalog_renderSnippet(t *testing.T) {
	c := NewCatalog("en")
	tmpl, err := template.New("").Funcs(c.Funcs()).Parse(`<html lang="{{ htmlLang }}">{{ t "home.welcome" "Dev" }}</html>`)
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}
	body := buf.String()
	if !strings.Contains(body, `lang="en"`) {
		t.Errorf("body = %q", body)
	}
	if !strings.Contains(body, "Welcome, Dev!") {
		t.Errorf("body = %q", body)
	}
}
