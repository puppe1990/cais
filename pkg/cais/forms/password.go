package forms

import (
	"html/template"
	"strings"
)

const (
	passwordEyeShow = `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>`
	passwordEyeHide = `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>`
)

// FieldPassword renders a password input with a show/hide toggle button.
func FieldPassword(f FieldData) template.HTML {
	var b strings.Builder
	b.WriteString(`<div><label class="block text-sm font-medium text-slate-700 mb-1" for="`)
	b.WriteString(template.HTMLEscapeString(f.Name))
	b.WriteString(`">`)
	b.WriteString(template.HTMLEscapeString(f.Label))
	b.WriteString(`</label><div class="cais-password-wrap"><input class="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-indigo-500 outline-none" type="password" id="`)
	b.WriteString(template.HTMLEscapeString(f.Name))
	b.WriteString(`" name="`)
	b.WriteString(template.HTMLEscapeString(f.Name))
	b.WriteString(`"`)
	if f.Required {
		b.WriteString(` required`)
	}
	b.WriteString(` /><button type="button" class="cais-password-toggle" data-cais-password-toggle aria-label="Show password"><span data-cais-password-icon="show">`)
	b.WriteString(passwordEyeShow)
	b.WriteString(`</span><span data-cais-password-icon="hide" class="hidden">`)
	b.WriteString(passwordEyeHide)
	b.WriteString(`</span></button></div>`)
	if f.Error != "" {
		b.WriteString(`<p class="text-red-600 text-sm mt-1">`)
		b.WriteString(template.HTMLEscapeString(f.Error))
		b.WriteString(`</p>`)
	}
	b.WriteString(`</div>`)
	return template.HTML(b.String())
}
