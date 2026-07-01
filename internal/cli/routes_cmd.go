package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RouteEntry is a single HTTP route parsed from routes.go.
type RouteEntry struct {
	Method string
	Path   string
}

var routePattern = regexp.MustCompile(`(?:r|g)\.(Get|Post|Put|Patch|Delete)\("([^"]+)"`)

func parseRoutesFile(path string) ([]RouteEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseRoutesContent(string(data)), nil
}

func parseRoutesContent(content string) []RouteEntry {
	matches := routePattern.FindAllStringSubmatch(content, -1)
	entries := make([]RouteEntry, 0, len(matches))
	for _, m := range matches {
		entries = append(entries, RouteEntry{
			Method: strings.ToUpper(m[1]),
			Path:   m[2],
		})
	}
	return entries
}

func formatRoutes(entries []RouteEntry) string {
	lines := make([]string, len(entries))
	for i, e := range entries {
		lines[i] = formatRouteEntry(e)
	}
	return strings.Join(lines, "\n")
}

func formatRouteEntry(e RouteEntry) string {
	return fmt.Sprintf("%-4s %s", e.Method, e.Path)
}

func (c *CLI) cmdRoutes() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "internal/app/routes.go")
	entries, err := parseRoutesFile(path)
	if err != nil {
		return fmt.Errorf("read routes: %w", err)
	}
	for _, e := range entries {
		_, _ = fmt.Fprintf(c.Out, "%s\n", formatRouteEntry(e))
	}
	return nil
}
