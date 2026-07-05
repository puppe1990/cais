package forms

import (
	"html/template"
	"strconv"
	"strings"
)

// SelectOption is one entry in a foreign-key select field.
type SelectOption struct {
	Value string
	Label string
}

// SelectFieldData describes a labeled select for templates.
type SelectFieldData struct {
	Name     string
	Label    string
	Value    string
	Required bool
	Error    string
	Options  []SelectOption
}

// MakeSelectField builds SelectFieldData from a selected int64 FK value.
func MakeSelectField(name, label string, value int64, options []SelectOption, required bool, errors map[string]string) SelectFieldData {
	val := ""
	if value != 0 {
		val = strconv.FormatInt(value, 10)
	}
	return SelectFieldData{
		Name:     name,
		Label:    label,
		Value:    val,
		Required: required,
		Error:    FieldError(errors, name),
		Options:  options,
	}
}

// MakeSelectFieldPtr builds SelectFieldData from an optional int64 FK pointer.
func MakeSelectFieldPtr(name, label string, value *int64, options []SelectOption, required bool, errors map[string]string) SelectFieldData {
	val := ""
	if value != nil && *value != 0 {
		val = strconv.FormatInt(*value, 10)
	}
	return SelectFieldData{
		Name:     name,
		Label:    label,
		Value:    val,
		Required: required,
		Error:    FieldError(errors, name),
		Options:  options,
	}
}

// FieldSelect renders a labeled select with optional error text.
func FieldSelect(f SelectFieldData) template.HTML {
	var b strings.Builder
	b.WriteString(`<div><label class="block text-sm font-medium text-slate-700 mb-1" for="`)
	b.WriteString(template.HTMLEscapeString(f.Name))
	b.WriteString(`">`)
	b.WriteString(template.HTMLEscapeString(f.Label))
	b.WriteString(`</label><select data-cais-select-search="true" class="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-indigo-500 outline-none" id="`)
	b.WriteString(template.HTMLEscapeString(f.Name))
	b.WriteString(`" name="`)
	b.WriteString(template.HTMLEscapeString(f.Name))
	b.WriteString(`"`)
	if f.Required {
		b.WriteString(` required`)
	}
	b.WriteString(`>`)
	if f.Required {
		b.WriteString(`<option value="">Select `)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</option>`)
	} else {
		b.WriteString(`<option value="">—</option>`)
	}
	for _, opt := range f.Options {
		b.WriteString(`<option value="`)
		b.WriteString(template.HTMLEscapeString(opt.Value))
		b.WriteString(`"`)
		if opt.Value == f.Value {
			b.WriteString(` selected`)
		}
		b.WriteString(`>`)
		b.WriteString(template.HTMLEscapeString(opt.Label))
		b.WriteString(`</option>`)
	}
	b.WriteString(`</select>`)
	if f.Error != "" {
		b.WriteString(`<p class="text-red-600 text-sm mt-1">`)
		b.WriteString(template.HTMLEscapeString(f.Error))
		b.WriteString(`</p>`)
	}
	b.WriteString(`</div>`)
	return template.HTML(b.String())
}
