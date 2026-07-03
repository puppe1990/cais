// Resource admin/public HTML template generation for cais g resource.
package cli

import (
	"fmt"
	"strings"
)

func buildAdminFormHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		switch f.Widget {
		case "select":
			if f.GoType == "*int64" {
				fmt.Fprintf(&fields, `    {{ fieldSelect (makeSelectFieldPtr "%s" "%s" .Item.%s .%sOptions %t .Errors) }}
`, f.Name, f.RefPascal, f.Pascal, f.RefPascal, f.Required)
			} else {
				fmt.Fprintf(&fields, `    {{ fieldSelect (makeSelectField "%s" "%s" .Item.%s .%sOptions %t .Errors) }}
`, f.Name, f.RefPascal, f.Pascal, f.RefPascal, f.Required)
			}
		case "textarea":
			fmt.Fprintf(&fields, `    {{ fieldInput (makeField "%s" "%s" .Item.%s "textarea" %t .Errors) }}
`, f.Name, f.Pascal, f.Pascal, f.Required)
		case "checkbox":
			fmt.Fprintf(&fields, `    <label class="flex items-center gap-2 text-sm text-slate-700">
      <input type="checkbox" name="%s" class="rounded border-slate-300 text-indigo-600" {{ if .Item.%s }}checked{{ end }} />
      %s
    </label>
`, f.Name, f.Pascal, f.Pascal)
		default:
			fmt.Fprintf(&fields, `    {{ fieldInput (makeField "%s" "%s" .Item.%s "%s" %t .Errors) }}
`, f.Name, f.Pascal, f.Pascal, f.HTMLType, f.Required)
		}
	}
	return fmt.Sprintf(`{{ define "title" }}{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}{{ end }} {{ define "content" }}
<div class="max-w-md mx-auto">
  <a href="/admin/%s" class="text-sm text-indigo-600 hover:underline mb-4 inline-block">← Back</a>
  <h1 class="text-3xl font-bold text-slate-900 mb-6">{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}</h1>
  <form id="admin-%s-form" class="bg-white rounded-2xl border border-slate-200 p-6 shadow-sm space-y-4" method="post"
    action="{{ if .IsNew }}/admin/%s{{ else }}/admin/%s/{{ .Item.ID }}{{ end }}"
    {{ if .IsNew }}{{ hxForm "/admin/%s" "#admin-%s-errors" "" }}{{ else }}{{ hxForm (printf "/admin/%s/%%d" .Item.ID) "#admin-%s-errors" "" }}{{ end }}>
    <div id="admin-%s-errors"></div>
    {{ csrfField .CSRFToken }}
%s
    <button type="submit" class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition">
      {{ if .IsNew }}Create{{ else }}Save{{ end }}
    </button>
  </form>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Title, data.Title, data.Snake, data.Plural, data.Plural, data.Plural, data.Snake, data.Plural, data.Snake, data.Snake, fields.String())
}

func buildAdminFormErrorsPartial(data scaffoldData) string {
	return fmt.Sprintf(`{{- define "admin_%s_form_errors" -}}
{{ range $field, $msg := .Errors }}
<p class="text-red-600 text-sm mb-2">{{ $msg }}</p>
{{ end }}
{{- end -}}
`, data.Snake)
}

func adminIndexDisplayField(fields []FieldDef) FieldDef {
	displayField := fields[0]
	for _, f := range fields {
		if f.Name == "title" || f.Name == "name" {
			return f
		}
	}
	return displayField
}

func buildAdminPaginationBlock(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return fmt.Sprintf(`    <div class="flex items-center justify-between px-6 py-4 border-t bg-slate-50">
      {{ if .HasPrev }}
      <a href="/admin/%s?page={{ .PrevPage }}" {{ hxPaginate (printf "/admin/%s?page=%%d" .PrevPage) "#admin-%s" }} class="text-indigo-600 hover:underline">← Previous</a>
      {{ else }}
      <span></span>
      {{ end }}
      <span class="text-sm text-slate-500">Page {{ .Page }}</span>
      {{ if .HasNext }}
      <a href="/admin/%s?page={{ .NextPage }}" {{ hxPaginate (printf "/admin/%s?page=%%d" .NextPage) "#admin-%s" }} class="text-indigo-600 hover:underline">Next →</a>
      {{ else }}
      <span></span>
      {{ end }}
    </div>
`, data.Plural, data.Plural, data.Plural, data.Plural, data.Plural, data.Plural)
}

func buildAdminIndexPanel(data scaffoldData) string {
	displayField := adminIndexDisplayField(data.Fields)
	return fmt.Sprintf(`    <table class="w-full text-left text-sm">
      <thead class="bg-slate-50 border-b"><tr><th class="px-6 py-3">%s</th><th class="px-6 py-3 text-right">Actions</th></tr></thead>
      <tbody class="divide-y">
        {{ range .Items }}
        <tr class="hover:bg-slate-50">
          <td class="px-6 py-4 font-medium">{{ .%s }}</td>
          <td class="px-6 py-4 text-right space-x-3">
            <a href="/admin/%s/{{ .ID }}" class="text-indigo-600 hover:underline">View</a>
            <a href="/admin/%s/{{ .ID }}/edit" class="text-slate-600 hover:underline">Edit</a>
            <button type="button" class="text-red-600 hover:underline" hx-post="/admin/%s/{{ .ID }}/delete" hx-target="closest tr" hx-swap="delete" hx-confirm="Delete this item?" data-cais-optimistic="remove">Delete</button>
          </td>
        </tr>
        {{ else }}
        <tr><td colspan="2" class="px-6 py-8 text-center text-slate-500">No items yet.</td></tr>
        {{ end }}
      </tbody>
    </table>
%s`, displayField.Pascal, displayField.Pascal, data.Plural, data.Plural, data.Plural, buildAdminPaginationBlock(data))
}

func buildAdminIndexPartial(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return fmt.Sprintf(`{{- define "admin_%s_index" -}}
%s
{{- end -}}
`, data.Plural, buildAdminIndexPanel(data))
}

func buildAdminIndexHTML(data scaffoldData) string {
	panel := buildAdminIndexPanel(data)
	if data.Paginate {
		panel = fmt.Sprintf(`{{ template "admin_%s_index" . }}`, data.Plural)
	}
	return fmt.Sprintf(`{{ define "title" }}Admin — %s{{ end }} {{ define "content" }}
<div class="max-w-3xl mx-auto">
  <div class="flex items-center justify-between mb-8">
    <h1 class="text-3xl font-bold text-slate-900">%s</h1>
    <a href="/admin/%s/new" class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition shadow-sm">+ New</a>
  </div>
  <div id="admin-%s" class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
%s
  </div>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Plural, panel)
}

func buildAdminShowHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		if f.GoType == "bool" {
			fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm font-medium text-slate-500">%s</dt>
      <dd class="mt-1 text-slate-900">{{ if .Item.%s }}Yes{{ else }}No{{ end }}</dd>
    </div>
`, f.Pascal, f.Pascal)
			continue
		}
		fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm font-medium text-slate-500">%s</dt>
      <dd class="mt-1 text-slate-900">{{ .Item.%s }}</dd>
    </div>
`, f.Pascal, f.Pascal)
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="max-w-md mx-auto">
  <a href="/admin/%s" class="text-sm text-indigo-600 hover:underline mb-4 inline-block">← Back</a>
  <h1 class="text-3xl font-bold text-slate-900 mb-6">%s</h1>
  <dl class="bg-white rounded-2xl border border-slate-200 p-6 shadow-sm space-y-4">
%s  </dl>
  <div class="mt-6 flex gap-3">
    <a href="/admin/%s/{{ .Item.ID }}/edit" class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 px-4 rounded-xl transition shadow-sm">Edit</a>
    <button type="button" class="text-red-600 hover:underline py-2 px-4" hx-post="/admin/%s/{{ .Item.ID }}/delete" hx-target="closest .max-w-md" hx-swap="delete" hx-confirm="Delete this item?" data-cais-optimistic="remove">Delete</button>
  </div>
</div>
{{ end }}
`, data.Title, data.Plural, data.Title, fields.String(), data.Plural, data.Plural)
}

func buildPublicListItemHTML(data scaffoldData) string {
	display := displayFieldForList(data.Fields)
	var linkField *FieldDef
	for i, f := range data.Fields {
		if f.HTMLType == "url" {
			linkField = &data.Fields[i]
			break
		}
	}

	var b strings.Builder
	if linkField != nil {
		fmt.Fprintf(&b, `<a href="{{ .%s }}" target="_blank" rel="noopener" class="text-lg font-semibold text-indigo-600 hover:underline">{{ .%s }}</a>`, linkField.Pascal, display.Pascal)
	} else {
		fmt.Fprintf(&b, `<p class="text-lg font-semibold text-slate-800">{{ .%s }}</p>`, display.Pascal)
	}

	var meta []string
	for _, f := range data.Fields {
		if f.Pascal == display.Pascal {
			continue
		}
		switch f.GoType {
		case "bool":
			meta = append(meta, fmt.Sprintf(`<span hx-post="/%s/{{ .ID }}/toggle" {{ hxMorphOuter }} hx-target="this" data-cais-optimistic="toggle" class="cursor-pointer inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {{ if .%s }}bg-green-50 text-green-700{{ else }}bg-slate-100 text-slate-600{{ end }}">{{ if .%s }}%s{{ else }}Pending{{ end }}</span>`, data.Plural, f.Pascal, f.Pascal, f.Pascal))
		case "int64", "*int64", "float64", "*float64":
			meta = append(meta, fmt.Sprintf(`<span class="text-sm text-slate-500">%s: {{ .%s }}</span>`, f.Pascal, f.Pascal))
		}
	}
	if len(meta) > 0 {
		b.WriteString(`<div class="mt-2 flex flex-wrap items-center gap-2">`)
		b.WriteString(strings.Join(meta, "\n"))
		b.WriteString(`</div>`)
	}

	for _, f := range data.Fields {
		if f.Pascal == display.Pascal || f.Widget != "textarea" {
			continue
		}
		fmt.Fprintf(&b, `{{ if .%s }}<p class="mt-2 text-sm text-slate-600 line-clamp-2">{{ .%s }}</p>{{ end }}`, f.Pascal, f.Pascal)
	}

	return b.String()
}

func buildPublicPaginationBlock(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return fmt.Sprintf(`  <div class="flex items-center justify-between mt-6">
    {{ if .HasPrev }}
    <a href="/%s?page={{ .PrevPage }}" {{ hxPaginate (printf "/%s?page=%%d" .PrevPage) "#%s-panel" }} class="text-indigo-600 hover:underline">← Previous</a>
    {{ else }}
    <span></span>
    {{ end }}
    <span class="text-sm text-slate-500">Page {{ .Page }}</span>
    {{ if .HasNext }}
    <a href="/%s?page={{ .NextPage }}" {{ hxPaginate (printf "/%s?page=%%d" .NextPage) "#%s-panel" }} class="text-indigo-600 hover:underline">Next →</a>
    {{ else }}
    <span></span>
    {{ end }}
  </div>
`, data.Plural, data.Plural, data.Plural, data.Plural, data.Plural, data.Plural)
}

func buildPublicListPanel(data scaffoldData) string {
	itemBlock := buildPublicListItemHTML(data)
	return fmt.Sprintf(`  <ul id="%s-list" class="space-y-3">
    {{ range .Items }}
    <li class="bg-white rounded-2xl border border-slate-200 p-5 shadow-sm">%s</li>
    {{ else }}
    <li class="text-center text-slate-500 py-8">No items yet.</li>
    {{ end }}
  </ul>
%s`, data.Plural, itemBlock, buildPublicPaginationBlock(data))
}

func buildPublicListPartial(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return fmt.Sprintf(`{{- define "%s_list" -}}
%s
{{- end -}}
`, data.Plural, buildPublicListPanel(data))
}

func buildPublicListHTML(data scaffoldData) string {
	pluralTitle := toTitle(data.Plural)
	intField := firstIntField(data.Fields)
	totalBlock := ""
	if intField != nil {
		totalVar := "Total"
		if data.Paginate {
			totalVar = "Sum"
		}
		totalBlock = fmt.Sprintf(`  <div class="bg-white rounded-2xl border border-slate-200 p-4 shadow-sm mb-6 flex items-center justify-between">
    <span class="text-sm font-medium text-slate-500">Total %s</span>
    <span class="text-2xl font-bold text-indigo-600">{{ .%s }}</span>
  </div>
`, intField.Pascal, totalVar)
	}
	listBlock := buildPublicListPanel(data)
	if data.Paginate {
		listBlock = fmt.Sprintf(`  <div id="%s-panel">
{{ template "%s_list" . }}
  </div>`, data.Plural, data.Plural)
	} else {
		listBlock = "  " + listBlock
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="max-w-2xl mx-auto">
  <h1 class="text-3xl font-bold text-slate-900 mb-6">%s</h1>
%s%s
</div>
{{ end }}
`, pluralTitle, pluralTitle, totalBlock, listBlock)
}

func buildPublicTogglePartial(data scaffoldData) string {
	boolField := firstBoolField(data.Fields)
	if boolField == nil {
		return ""
	}
	return fmt.Sprintf(`{{- define "%s_toggle" -}}<span hx-post="/%s/{{ .ID }}/toggle" {{ hxMorphOuter }} hx-target="this" data-cais-optimistic="toggle" class="cursor-pointer inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {{ if .%s }}bg-green-50 text-green-700{{ else }}bg-slate-100 text-slate-600{{ end }}">{{ if .%s }}%s{{ else }}Pending{{ end }}</span>{{- end -}}
`, data.Plural, data.Plural, boolField.Pascal, boolField.Pascal, boolField.Pascal)
}
