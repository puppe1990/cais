package cli

import "strings"

func nullableStoreHelpers(fields []FieldDef) string {
	var b strings.Builder
	if fieldNeedsStrPtr(fields) {
		b.WriteString(`
func strPtr(s string) *string { return &s }
`)
	}
	if fieldNeedsInt64Ptr(fields) {
		b.WriteString(`
func int64Ptr(n int64) *int64 { return &n }
`)
	}
	return b.String()
}

func seedValueForField(f FieldDef) string {
	name := strings.ToLower(f.Name)
	switch f.GoType {
	case "*string":
		return "strPtr(" + seedStringValue(f, name) + ")"
	case "*int64":
		return "int64Ptr(" + seedIntValue(f, name) + ")"
	case "bool":
		if strings.Contains(name, "active") || strings.Contains(name, "enabled") {
			return "true"
		}
		return "false"
	case "int64":
		return seedIntValue(f, name)
	default:
		return seedStringValue(f, name)
	}
}

func seedIntValue(f FieldDef, name string) string {
	if strings.Contains(name, "count") || strings.Contains(name, "total") {
		return "10"
	}
	if strings.Contains(name, "price") || strings.Contains(name, "amount") {
		return "99"
	}
	if strings.Contains(name, "age") || strings.Contains(name, "year") {
		return "25"
	}
	if strings.Contains(name, "rating") || strings.Contains(name, "score") {
		return "5"
	}
	if strings.Contains(name, "minute") || strings.Contains(name, "hour") || strings.Contains(name, "second") || strings.Contains(name, "duration") || strings.Contains(name, "calorie") {
		return "30"
	}
	if strings.Contains(name, "quantity") || strings.Contains(name, "qty") || strings.Contains(name, "servings") {
		return "4"
	}
	return "1"
}

func seedStringValue(f FieldDef, name string) string {
	if f.HTMLType == "url" {
		if strings.Contains(name, "github") {
			return `"https://github.com/example"`
		}
		if strings.Contains(name, "twitter") || strings.Contains(name, "x") {
			return `"https://twitter.com/example"`
		}
		return `"https://example.com"`
	}
	if f.Widget == "textarea" {
		if strings.Contains(name, "description") {
			return `"A detailed description of this item."`
		}
		if strings.Contains(name, "notes") || strings.Contains(name, "comment") {
			return `"Some notes about this entry."`
		}
		return `"Lorem ipsum dolor sit amet, consectetur adipiscing elit."`
	}
	if f.HTMLType == "date" {
		return `"2024-01-15"`
	}
	if strings.Contains(name, "email") {
		return `"user@example.com"`
	}
	if strings.Contains(name, "name") || strings.Contains(name, "title") {
		return `"Sample Item"`
	}
	if strings.Contains(name, "status") {
		return `"active"`
	}
	if strings.Contains(name, "category") {
		return `"general"`
	}
	return `"Sample"`
}

func insertColumns(fields []FieldDef) (cols, placeholders string) {
	names := fieldNames(fields)
	ph := make([]string, len(names))
	for i := range names {
		ph[i] = "?"
	}
	return strings.Join(names, ", "), strings.Join(ph, ", ")
}

func insertArgs(fields []FieldDef) string {
	var args []string
	for _, f := range fields {
		if f.GoType == "bool" {
			args = append(args, "boolInt(c."+f.Pascal+")")
		} else {
			args = append(args, "c."+f.Pascal)
		}
	}
	return strings.Join(args, ", ")
}

func updateSets(fields []FieldDef) string {
	var sets []string
	for _, f := range fields {
		sets = append(sets, f.Name+" = ?")
	}
	return strings.Join(sets, ", ")
}

func selectColumns(fields []FieldDef) string {
	return strings.Join(fieldNames(fields), ", ")
}

func fieldNames(fields []FieldDef) []string {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names
}

func boolScanTemp(f FieldDef) string {
	return f.Name + "Int"
}

func scanDeclare(fields []FieldDef) string {
	var extra []string
	for _, f := range fields {
		if f.GoType == "bool" {
			extra = append(extra, "\tvar "+boolScanTemp(f)+" int")
		}
	}
	if len(extra) == 0 {
		return ""
	}
	return strings.Join(extra, "\n") + "\n"
}

func scanLoopDeclare(fields []FieldDef) string {
	var extra []string
	for _, f := range fields {
		if f.GoType == "bool" {
			extra = append(extra, "\t\tvar "+boolScanTemp(f)+" int")
		}
	}
	if len(extra) == 0 {
		return ""
	}
	return strings.Join(extra, "\n") + "\n"
}

func scanVars(fields []FieldDef) string {
	var vars []string
	for _, f := range fields {
		if f.GoType == "bool" {
			vars = append(vars, "&"+boolScanTemp(f))
		} else {
			vars = append(vars, "&c."+f.Pascal)
		}
	}
	vars = append(vars, "&c.CreatedAt")
	return "&c.ID, " + strings.Join(vars, ", ")
}

func scanAssign(fields []FieldDef) string {
	var lines []string
	for _, f := range fields {
		if f.GoType == "bool" {
			lines = append(lines, "\tc."+f.Pascal+" = "+boolScanTemp(f)+" == 1")
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func scanLoopAssign(fields []FieldDef) string {
	var lines []string
	for _, f := range fields {
		if f.GoType == "bool" {
			lines = append(lines, "\t\tc."+f.Pascal+" = "+boolScanTemp(f)+" == 1")
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func needsStrconv(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "int64" || f.GoType == "*int64" {
			return true
		}
	}
	return false
}

func hasBoolField(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "bool" {
			return true
		}
	}
	return false
}

func boolImport(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}

func firstBoolField(fields []FieldDef) *FieldDef {
	for i, f := range fields {
		if f.GoType == "bool" {
			return &fields[i]
		}
	}
	return nil
}

func firstIntField(fields []FieldDef) *FieldDef {
	for i, f := range fields {
		if f.Widget == "select" {
			continue
		}
		if f.GoType == "int64" || f.GoType == "*int64" {
			return &fields[i]
		}
	}
	return nil
}

func sumArg(intField *FieldDef) string {
	if intField == nil {
		return ""
	}
	return ", Total: total"
}

func displayFieldForList(fields []FieldDef) FieldDef {
	for _, f := range fields {
		if f.Name == "title" || f.Name == "name" {
			return f
		}
	}
	return fields[0]
}
