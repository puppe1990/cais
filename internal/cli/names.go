package cli

import (
	"strings"
	"unicode"
)

type scaffoldData struct {
	AppName    string
	ModulePath string
	Handler    string
	Pascal     string
	Camel      string
	Snake      string
	Title      string
}

func dataForHandler(name string) scaffoldData {
	pascal := toPascal(name)
	return scaffoldData{
		Handler: name,
		Pascal:  pascal,
		Camel:   lowerFirst(pascal),
		Snake:   toSnake(name),
		Title:   toTitle(name),
	}
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
