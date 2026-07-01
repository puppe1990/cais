package sqllog

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var fromTableRE = regexp.MustCompile(`(?i)\bfrom\s+([a-z_][a-z0-9_]*)`)
var intoTableRE = regexp.MustCompile(`(?i)\binto\s+([a-z_][a-z0-9_]*)`)
var updateTableRE = regexp.MustCompile(`(?i)\bupdate\s+([a-z_][a-z0-9_]*)`)

func operationLabel(query string) string {
	q := strings.TrimSpace(query)
	upper := strings.ToUpper(q)
	table := extractTable(q)

	switch {
	case strings.HasPrefix(upper, "SELECT"):
		if table != "" {
			return singularize(table) + " Load"
		}
	case strings.HasPrefix(upper, "INSERT"):
		if table != "" {
			return singularize(table) + " Create"
		}
	case strings.HasPrefix(upper, "UPDATE"):
		if table != "" {
			return singularize(table) + " Update"
		}
	case strings.HasPrefix(upper, "DELETE"):
		if table != "" {
			return singularize(table) + " Destroy"
		}
	}
	return "SQL"
}

func extractTable(query string) string {
	q := strings.TrimSpace(query)
	upper := strings.ToUpper(q)

	switch {
	case strings.HasPrefix(upper, "SELECT"):
		if m := fromTableRE.FindStringSubmatch(q); len(m) == 2 {
			return m[1]
		}
	case strings.HasPrefix(upper, "INSERT"):
		if m := intoTableRE.FindStringSubmatch(q); len(m) == 2 {
			return m[1]
		}
	case strings.HasPrefix(upper, "UPDATE"):
		if m := updateTableRE.FindStringSubmatch(q); len(m) == 2 {
			return m[1]
		}
	case strings.HasPrefix(upper, "DELETE"):
		if m := fromTableRE.FindStringSubmatch(q); len(m) == 2 {
			return m[1]
		}
	}
	return ""
}

func singularize(table string) string {
	table = strings.ToLower(strings.Trim(table, `"`))
	switch {
	case strings.HasSuffix(table, "ies"):
		table = table[:len(table)-3] + "y"
	case strings.HasSuffix(table, "ses"):
		table = table[:len(table)-1]
	case strings.HasSuffix(table, "s") && !strings.HasSuffix(table, "ss"):
		table = table[:len(table)-1]
	}

	parts := strings.Split(table, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func formatArgs(args []any) string {
	if len(args) == 0 {
		return "[]"
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		if s, ok := arg.(string); ok {
			parts[i] = fmt.Sprintf("%q", s)
			continue
		}
		parts[i] = fmt.Sprint(arg)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	case d < time.Second:
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
