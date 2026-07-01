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

func TestEmail_invalid(t *testing.T) {
	for _, addr := range []string{"", "not-an-email", "@missing.com", "user@"} {
		if err := Email(addr); err == nil {
			t.Errorf("Email(%q) = nil, want error", addr)
		}
	}
}
