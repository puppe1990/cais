package forms

import (
	"html/template"
	"strings"
)

// FieldData describes a single form field for templates.
type FieldData struct {
	Name     string
	Label    string
	Value    string
	Type     string
	Required bool
	Error    string
}

// MakeField builds FieldData with validation error from a map.
func MakeField(name, label, value, htmlType string, required bool, errors map[string]string) FieldData {
	if htmlType == "" {
		htmlType = "text"
	}
	return FieldData{
		Name:     name,
		Label:    label,
		Value:    value,
		Type:     htmlType,
		Required: required,
		Error:    FieldError(errors, name),
	}
}

// FieldInput renders a labeled input, textarea, or checkbox with optional error text.
func FieldInput(f FieldData) template.HTML {
	var b strings.Builder
	switch f.Type {
	case "textarea":
		b.WriteString(`<div><label class="block text-sm font-medium text-slate-700 mb-1" for="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</label><textarea class="w-full border border-slate-300 rounded-lg px-3 py-2 min-h-[80px] focus:ring-2 focus:ring-indigo-500 outline-none" id="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" name="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(f.Value))
		b.WriteString(`</textarea>`)
	case "checkbox":
		b.WriteString(`<label class="flex items-center gap-2 text-sm text-slate-700"><input type="checkbox" name="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" class="rounded border-slate-300 text-indigo-600"`)
		if f.Value == "true" || f.Value == "1" || strings.EqualFold(f.Value, "on") {
			b.WriteString(` checked`)
		}
		b.WriteString(` />`)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</label>`)
	default:
		b.WriteString(`<div><label class="block text-sm font-medium text-slate-700 mb-1" for="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</label><input class="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-indigo-500 outline-none" type="`)
		b.WriteString(template.HTMLEscapeString(f.Type))
		b.WriteString(`" id="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" name="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(template.HTMLEscapeString(f.Value))
		b.WriteString(`"`)
		if f.Required {
			b.WriteString(` required`)
		}
		b.WriteString(` />`)
	}
	if f.Error != "" && f.Type != "checkbox" {
		b.WriteString(`<p class="text-red-600 text-sm mt-1">`)
		b.WriteString(template.HTMLEscapeString(f.Error))
		b.WriteString(`</p>`)
	}
	if f.Type != "checkbox" {
		b.WriteString(`</div>`)
	}
	return template.HTML(b.String())
}
