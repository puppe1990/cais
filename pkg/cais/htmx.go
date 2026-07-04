package cais

import (
	"fmt"
	"net/http"
	"strings"
)

func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// SetTrigger sets HX-Trigger for client-side events after swap.
func SetTrigger(w http.ResponseWriter, id string) {
	w.Header().Set("HX-Trigger", id)
}

// SetToast sets HX-Trigger so cais.js shows a transient toast after the HTMX swap.
// Non-ASCII runes are escaped as \uXXXX so browsers read the header without mojibake.
func SetToast(w http.ResponseWriter, message string) {
	w.Header().Set("HX-Trigger", hxTriggerPayload("caisToast", message))
}

// SetRetarget sets HX-Retarget to change the swap target from the response.
func SetRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}

// SetFocus sets HX-Trigger so cais.js focuses a field after swap (e.g. first invalid input).
func SetFocus(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Trigger", hxTriggerPayload("caisFocus", selector))
}

func hxTriggerPayload(key, value string) string {
	return fmt.Sprintf(`{%s:%s}`, jsonStringASCII(key), jsonStringASCII(value))
}

func jsonStringASCII(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r > 0x7e {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
