package cli

import (
	"fmt"
	"strings"
)

type FieldDef struct {
	Name     string
	Pascal   string
	SQLType  string
	GoType   string
	HTMLType string
	Widget   string
	Required bool
}

func parseFields(spec string) ([]FieldDef, error) {
	if strings.TrimSpace(spec) == "" {
		return defaultFields(), nil
	}

	var fields []FieldDef
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, typ, req, err := parseFieldPart(part)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldFromNameType(name, typ, req))
	}
	if len(fields) == 0 {
		return defaultFields(), nil
	}
	return fields, nil
}

func parseFieldPart(part string) (name, typ string, required bool, err error) {
	required = true
	if idx := strings.Index(part, ":"); idx >= 0 {
		name = strings.TrimSpace(part[:idx])
		rest := strings.TrimSpace(part[idx+1:])
		if strings.HasSuffix(rest, "?") {
			required = false
			rest = strings.TrimSuffix(rest, "?")
		}
		typ = rest
	} else {
		name = part
		typ = "string"
	}
	if name == "" {
		return "", "", false, fmt.Errorf("invalid field %q", part)
	}
	return name, typ, required, nil
}

func fieldFromNameType(name, typ string, required bool) FieldDef {
	pascal := toPascal(name)
	if toSnake(name) == "url" {
		pascal = "URL"
	}
	switch typ {
	case "text":
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "TEXT NOT NULL DEFAULT ''", GoType: "string", HTMLType: "text", Widget: "textarea", Required: required}
	case "url":
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "TEXT NOT NULL", GoType: "string", HTMLType: "url", Widget: "input", Required: required}
	case "bool":
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "INTEGER NOT NULL DEFAULT 0", GoType: "bool", HTMLType: "checkbox", Widget: "checkbox", Required: false}
	case "int":
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "INTEGER NOT NULL DEFAULT 0", GoType: "int64", HTMLType: "number", Widget: "input", Required: required}
	case "date":
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "TEXT NOT NULL DEFAULT ''", GoType: "string", HTMLType: "date", Widget: "input", Required: required}
	default:
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "TEXT NOT NULL", GoType: "string", HTMLType: "text", Widget: "input", Required: required}
	}
}

func defaultFields() []FieldDef {
	return []FieldDef{fieldFromNameType("name", "string", true)}
}
