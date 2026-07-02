package devlog

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatForDisplay renders buffered log lines for /logs: JSON entries become
// readable one-liners; legacy plain-text lines pass through unchanged.
// Stdout and the buffer keep raw JSON so agents can grep structured fields.
func FormatForDisplay(text string) string {
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return ""
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		if formatted, ok := formatJSONLine(line); ok {
			lines = append(lines, formatted)
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func formatJSONLine(line string) (string, bool) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return "", false
	}
	kind, _ := raw["kind"].(string)
	switch kind {
	case "request":
		method, _ := raw["method"].(string)
		path, _ := raw["path"].(string)
		phase, _ := raw["phase"].(string)
		remote, _ := raw["remote"].(string)
		switch phase {
		case "started":
			return fmt.Sprintf("→ %s %s (%s)", method, path, remote), true
		case "completed":
			status, _ := raw["status"].(float64)
			dur, _ := raw["duration_ms"].(float64)
			return fmt.Sprintf("← %d %s %s in %.1fms", int(status), method, path, dur), true
		}
	case "sql":
		op, _ := raw["operation"].(string)
		query, _ := raw["query"].(string)
		dur, _ := raw["duration_ms"].(float64)
		args := formatArgs(raw["args"])
		line := fmt.Sprintf("SQL %s (%.1fms) %s", op, dur, query)
		if args != "" {
			line += " " + args
		}
		if errMsg, _ := raw["error"].(string); errMsg != "" {
			line += " ERROR: " + errMsg
		}
		return line, true
	}
	return "", false
}

func formatArgs(v any) string {
	args, ok := v.([]any)
	if !ok || len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprint(arg)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
