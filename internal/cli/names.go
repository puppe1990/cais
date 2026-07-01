package cli

import (
	"strings"
	"unicode"
)

type scaffoldData struct {
	AppName      string
	ModulePath   string
	Handler      string
	Pascal       string
	Camel        string
	Snake        string
	Title        string
	Plural       string
	PluralPascal string
	PluralCamel  string
	MigrationNum string
}

func dataForHandler(name string) scaffoldData {
	pascal := toPascal(name)
	plural := toPlural(toSnake(name))
	pluralPascal := toPascal(plural)
	return scaffoldData{
		Handler:      name,
		Pascal:       pascal,
		Camel:        lowerFirst(pascal),
		Snake:        toSnake(name),
		Title:        toTitle(name),
		Plural:       plural,
		PluralPascal: pluralPascal,
		PluralCamel:  lowerFirst(pluralPascal),
	}
}

func dataForResource(name string) scaffoldData {
	return dataForHandler(name)
}

func toPlural(snake string) string {
	if strings.HasSuffix(snake, "s") {
		return snake
	}
	if strings.HasSuffix(snake, "y") && len(snake) > 1 {
		return snake[:len(snake)-1] + "ies"
	}
	return snake + "s"
}

func toPascal(s string) string {
	parts := splitName(s)
	for i, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}
	return strings.Join(parts, "")
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func toSnake(s string) string {
	return strings.Join(splitName(s), "_")
}

func toTitle(s string) string {
	parts := splitName(s)
	for i, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}
	return strings.Join(parts, " ")
}

func splitName(s string) []string {
	s = strings.ReplaceAll(s, "-", "_")
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == ' '
	})
}
