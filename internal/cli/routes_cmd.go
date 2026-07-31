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
	Method     string
	Path       string
	Handler    string
	Middleware string
}

var routePattern = regexp.MustCompile(`(?:r|g)\.(Get|Post|Put|Patch|Delete)\("([^"]+)"(?:,\s*(.+)\))?\s*$`)

func parseRoutesFile(path string) ([]RouteEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseRoutesContent(string(data)), nil
}

func parseRoutesContent(content string) []RouteEntry {
	return parseRoutesVerbose(content)
}

func parseRoutesVerbose(content string) []RouteEntry {
	lines := strings.Split(content, "\n")
	var entries []RouteEntry
	currentMiddleware := ""

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.Contains(trim, "r.Group(") || strings.Contains(trim, "g.Group(") {
			if idx := strings.Index(trim, "Group("); idx >= 0 {
				rest := trim[idx+6:]
				if end := strings.Index(rest, ","); end >= 0 {
					currentMiddleware = strings.TrimSpace(rest[:end])
				} else if end := strings.Index(rest, ")"); end >= 0 {
					currentMiddleware = strings.TrimSpace(rest[:end])
				}
			}
			continue
		}
		if trim == "})" || trim == "}" {
			currentMiddleware = ""
		}

		m := routePattern.FindStringSubmatch(trim)
		if len(m) < 3 {
			continue
		}
		handler := ""
		if len(m) >= 4 {
			handler = strings.TrimSpace(m[3])
		}
		entries = append(entries, RouteEntry{
			Method:     strings.ToUpper(m[1]),
			Path:       m[2],
			Handler:    handler,
			Middleware: currentMiddleware,
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

func formatRouteEntryVerbose(e RouteEntry) string {
	line := fmt.Sprintf("%-4s %-35s %s", e.Method, e.Path, e.Handler)
	if e.Middleware != "" {
		line += "  [" + e.Middleware + "]"
	}
	return strings.TrimRight(line, " ")
}

func (c *CLI) cmdRoutes(args []string) error {
	verbose := false
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		}
	}

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
		if verbose {
			_, _ = fmt.Fprintf(c.Out, "%s\n", formatRouteEntryVerbose(e))
		} else {
			_, _ = fmt.Fprintf(c.Out, "%s\n", formatRouteEntry(e))
		}
	}
	// Static conflict scan — same class of ServeMux panics that hit boot (#142).
	for _, warn := range routeConflictWarnings(entries) {
		_, _ = fmt.Fprintf(c.Out, "⚠ %s\n", warn)
	}
	return nil
}

// routeConflictWarnings returns human-readable ServeMux overlap warnings for parsed routes.
func routeConflictWarnings(entries []RouteEntry) []string {
	var warns []string
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			a, b := entries[i], entries[j]
			if a.Method != b.Method {
				continue
			}
			if !pathsConflictForServeMux(a.Path, b.Path) {
				continue
			}
			warns = append(warns, fmt.Sprintf(
				"possible ServeMux conflict: %s %s vs %s %s — prefer DELETE …/{id} or a distinct prefix (see AGENTS.md)",
				a.Method, a.Path, b.Method, b.Path,
			))
		}
	}
	return warns
}

// pathsConflictForServeMux mirrors pkg/cais patternsConflict (kept in CLI to avoid coupling parse to runtime).
func pathsConflictForServeMux(a, b string) bool {
	sa := splitRoutePath(a)
	sb := splitRoutePath(b)
	if len(sa) != len(sb) || len(sa) == 0 {
		return len(sa) == 0 && a == b
	}
	aMore, bMore := true, true
	for i := range sa {
		wa, wb := isRouteWild(sa[i]), isRouteWild(sb[i])
		switch {
		case !wa && !wb:
			if sa[i] != sb[i] {
				return false
			}
		case wa && !wb:
			aMore = false
		case !wa && wb:
			bMore = false
		default:
			aMore, bMore = false, false
		}
	}
	return aMore == bMore
}

func splitRoutePath(pattern string) []string {
	pattern = strings.Trim(pattern, "/")
	if pattern == "" {
		return nil
	}
	return strings.Split(pattern, "/")
}

func isRouteWild(seg string) bool {
	return strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
}
