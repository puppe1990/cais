package validate

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// MinLength checks trimmed s has at least min runes.
func MinLength(s string, min int) error {
	if utf8.RuneCountInString(strings.TrimSpace(s)) < min {
		return fmt.Errorf("must be at least %d characters", min)
	}
	return nil
}

// MaxLength checks trimmed s has at most max runes.
func MaxLength(s string, max int) error {
	if utf8.RuneCountInString(strings.TrimSpace(s)) > max {
		return fmt.Errorf("must be at most %d characters", max)
	}
	return nil
}
