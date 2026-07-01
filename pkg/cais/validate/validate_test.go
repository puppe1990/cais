package validate

import "testing"

func TestRequired(t *testing.T) {
	if err := Required(map[string]string{"name": "Ada"}, "name"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := Required(map[string]string{"name": "  "}, "name"); err == nil {
		t.Error("expected error for blank name")
	}
}

func TestURL(t *testing.T) {
	if err := URL("https://example.com"); err != nil {
		t.Errorf("URL() = %v", err)
	}
	if err := URL("not-a-url"); err == nil {
		t.Error("expected error for invalid url")
	}
}

func TestEmail_valid(t *testing.T) {
	for _, addr := range []string{"a@b.co", "user@example.com"} {
		if err := Email(addr); err != nil {
			t.Errorf("Email(%q) = %v, want nil", addr, err)
		}
	}
}

func TestInt(t *testing.T) {
	if err := Int("42"); err != nil {
		t.Errorf("Int(42) = %v", err)
	}
	for _, v := range []string{"", "abc", "3.14"} {
		if err := Int(v); err == nil {
			t.Errorf("Int(%q) = nil, want error", v)
		}
	}
}

func TestDate(t *testing.T) {
	if err := Date("2026-07-01"); err != nil {
		t.Errorf("Date() = %v", err)
	}
	for _, v := range []string{"", "01-07-2026", "2026-13-40"} {
		if err := Date(v); err == nil {
			t.Errorf("Date(%q) = nil, want error", v)
		}
	}
}

func TestEmail_invalid(t *testing.T) {
	for _, addr := range []string{"", "not-an-email", "@missing.com", "user@"} {
		if err := Email(addr); err == nil {
			t.Errorf("Email(%q) = nil, want error", addr)
		}
	}
}
