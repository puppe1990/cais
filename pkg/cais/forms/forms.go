package forms

import (
	"html/template"

	"github.com/puppe1990/cais/pkg/cais/csrf"
)

// Funcs returns template helpers for HTML forms.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"csrfField":          CSRFTokenField,
		"fieldError":         FieldError,
		"makeField":          MakeField,
		"fieldInput":         FieldInput,
		"fieldPassword":      FieldPassword,
		"makeSelectField":    MakeSelectField,
		"makeSelectFieldPtr": MakeSelectFieldPtr,
		"fieldSelect":        FieldSelect,
	}
}

// CSRFTokenField renders a hidden CSRF input for HTML forms.
func CSRFTokenField(token string) template.HTML {
	return template.HTML(csrf.FieldHTML(token))
}

// FieldError returns the validation message for a field, or empty when absent.
func FieldError(errors map[string]string, field string) string {
	if errors == nil {
		return ""
	}
	return errors[field]
}
