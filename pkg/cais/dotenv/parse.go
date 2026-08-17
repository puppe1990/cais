package dotenv

import "strings"

// Parse maps KEY=value lines from a .env file. Blank lines, comments, and
// lines without '=' are ignored. Process env override is applied by LoadFile.
func Parse(data []byte) map[string]string {
	vars := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		vars[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return vars
}
