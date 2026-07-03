package money

import "testing"

func TestFormatBRL_cents(t *testing.T) {
	tests := []struct {
		cents int
		want  string
	}{
		{549, "R$ 5,49"},
		{821, "R$ 8,21"},
		{2890, "R$ 28,90"},
		{0, "R$ 0,00"},
		{100, "R$ 1,00"},
	}
	for _, tc := range tests {
		if got := FormatBRL(tc.cents); got != tc.want {
			t.Errorf("FormatBRL(%d) = %q, want %q", tc.cents, got, tc.want)
		}
	}
}
