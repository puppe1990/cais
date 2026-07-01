package cli

import (
	"fmt"
	"strings"
)

// insertBeforeFunctionEnd finds funcName in content, locates its closing brace via
// brace-counting, and inserts insert immediately before that brace.
func insertBeforeFunctionEnd(content, funcName, insert string) (string, error) {
	needle := "func " + funcName + "("
	idx := strings.Index(content, needle)
	if idx == -1 {
		return "", fmt.Errorf("function %q not found", funcName)
	}

	rest := content[idx:]
	braceStart := strings.Index(rest, "{")
	if braceStart == -1 {
		return "", fmt.Errorf("opening brace for %q not found", funcName)
	}
	braceStart += idx

	depth := 0
	for i := braceStart; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[:i] + insert + content[i:], nil
			}
		}
	}
	return "", fmt.Errorf("closing brace for %q not found", funcName)
}
