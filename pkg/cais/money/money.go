package money

import "fmt"

// FormatBRL formats integer cents as Brazilian Real (e.g. 549 → "R$ 5,49").
func FormatBRL(cents int) string {
	if cents < 0 {
		cents = -cents
		return "-" + FormatBRL(cents)
	}
	reais := cents / 100
	frac := cents % 100
	return fmt.Sprintf("R$ %d,%02d", reais, frac)
}
