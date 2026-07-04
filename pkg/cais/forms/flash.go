package forms

import "github.com/puppe1990/cais/pkg/cais/flash"

// FlashMessage returns the human-readable flash text for layout templates.
// Use instead of {{ .Flash }}, which stringifies the struct as {notice Message}.
func FlashMessage(v any) string {
	switch m := v.(type) {
	case flash.Message:
		return m.Message
	case *flash.Message:
		if m == nil {
			return ""
		}
		return m.Message
	default:
		return ""
	}
}
