package i18n

import "testing"

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
