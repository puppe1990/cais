package validate

import "testing"

func TestMinLength(t *testing.T) {
	if err := MinLength("ab", 2); err != nil {
		t.Errorf("MinLength(ab, 2) = %v", err)
	}
	if err := MinLength("a", 2); err == nil {
		t.Error("expected error for short value")
	}
	if err := MinLength("  hi  ", 2); err != nil {
		t.Errorf("MinLength trims whitespace: %v", err)
	}
}

func TestMaxLength(t *testing.T) {
	if err := MaxLength("ab", 3); err != nil {
		t.Errorf("MaxLength(ab, 3) = %v", err)
	}
	if err := MaxLength("abcd", 3); err == nil {
		t.Error("expected error for long value")
	}
}

func TestMinLength_unicode(t *testing.T) {
	if err := MinLength("ação", 4); err != nil {
		t.Errorf("MinLength should count runes: %v", err)
	}
}
