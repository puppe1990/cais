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
		return defaultFields()
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
		f, err := fieldFromNameType(name, typ, req)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	if len(fields) == 0 {
		return defaultFields()
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

func fieldFromNameType(name, typ string, required bool) (FieldDef, error) {
	pascal := toPascal(name)
	if toSnake(name) == "url" {
		pascal = "URL"
	}
	switch typ {
	case "string":
		return stringField(name, pascal, required, "text", "input"), nil
	case "text":
		return stringField(name, pascal, required, "text", "textarea"), nil
	case "url":
		f := stringField(name, pascal, required, "url", "input")
		if required {
			f.SQLType = "TEXT NOT NULL"
		} else {
			f.SQLType = "TEXT"
		}
		return f, nil
	case "bool":
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "INTEGER NOT NULL DEFAULT 0", GoType: "bool", HTMLType: "checkbox", Widget: "checkbox", Required: false}, nil
	case "int":
		return intField(name, pascal, required), nil
	case "date":
		return stringField(name, pascal, required, "date", "input"), nil
	default:
		return FieldDef{}, fmt.Errorf("unknown field type %q (use string, text, url, bool, int, date)", typ)
	}
}

func stringField(name, pascal string, required bool, htmlType, widget string) FieldDef {
	f := FieldDef{
		Name:     toSnake(name),
		Pascal:   pascal,
		HTMLType: htmlType,
		Widget:   widget,
		Required: required,
	}
	if required {
		f.SQLType = "TEXT NOT NULL"
		if widget == "textarea" {
			f.SQLType = "TEXT NOT NULL DEFAULT ''"
		}
		f.GoType = "string"
	} else {
		f.SQLType = "TEXT"
		f.GoType = "*string"
	}
	return f
}

func intField(name, pascal string, required bool) FieldDef {
	f := FieldDef{
		Name:     toSnake(name),
		Pascal:   pascal,
		HTMLType: "number",
		Widget:   "input",
		Required: required,
	}
	if required {
		f.SQLType = "INTEGER NOT NULL DEFAULT 0"
		f.GoType = "int64"
	} else {
		f.SQLType = "INTEGER"
		f.GoType = "*int64"
	}
	return f
}

func defaultFields() ([]FieldDef, error) {
	f, err := fieldFromNameType("name", "string", true)
	if err != nil {
		return nil, err
	}
	return []FieldDef{f}, nil
}

func fieldNeedsStrPtr(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "*string" {
			return true
		}
	}
	return false
}

func fieldNeedsInt64Ptr(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "*int64" {
			return true
		}
	}
	return false
}
