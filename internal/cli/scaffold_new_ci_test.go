package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldNewApp_DropsUnusedBootstrap(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "noboot")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "noboot",
		ModulePath: "github.com/puppe1990/noboot",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "cmd/server/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "func bootstrap()") {
		t.Error("cmd/server/main.go must not define unused func bootstrap()")
	}
}

func TestScaffoldNewApp_ConsoleDoesNotFatalAfterDefer(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "consfatal")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "consfatal",
		ModulePath: "github.com/puppe1990/consfatal",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "cmd/console/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	src := string(body)
	runSrc := sourceFunc(src, "run")
	if runSrc == "" {
		t.Fatal("cmd/console/main.go must extract func run so defers run before exit")
	}
	if strings.Contains(runSrc, "log.Fatal") {
		t.Error("func run must not call log.Fatal (os.Exit skips defers)")
	}
	mainSrc := sourceFunc(src, "main")
	if strings.Contains(mainSrc, "defer ") && strings.Contains(mainSrc, "log.Fatal") {
		t.Error("func main must not call log.Fatal after defer")
	}
}

func TestScaffoldNewApp_ViteJSConfigsMatchPrettier(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "fmtjs")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "fmtjs",
		ModulePath: "github.com/puppe1990/fmtjs",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		rel     string
		needles []string
	}{
		{"vite.config.js", []string{`from "vite"`, `";`, `from "@sveltejs/vite-plugin-svelte"`}},
		{"svelte.config.js", []string{`from "@sveltejs/vite-plugin-svelte"`, `";`}},
		{"vitest-setup.js", []string{`import "@testing-library/jest-dom/vitest";`}},
	}
	for _, tc := range checks {
		body, err := os.ReadFile(filepath.Join(appDir, tc.rel))
		if err != nil {
			t.Errorf("%s: %v", tc.rel, err)
			continue
		}
		src := string(body)
		if strings.Contains(src, "'") {
			t.Errorf("%s: single-quoted strings fail Prettier (singleQuote: false)", tc.rel)
		}
		for _, needle := range tc.needles {
			if !strings.Contains(src, needle) {
				t.Errorf("%s missing %q", tc.rel, needle)
			}
		}
	}
}

func TestScaffoldNewApp_GoImportsLocalPrefixOrder(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	module := "github.com/puppe1990/imporder"
	appDir := filepath.Join(t.TempDir(), "imporder")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "imporder",
		ModulePath: module,
	}, false, false); err != nil {
		t.Fatal(err)
	}

	files := []string{
		"cmd/server/main.go",
		"internal/app/app.go",
		"internal/app/routes.go",
		"cmd/console/main.go",
		"internal/store/store.go",
		"internal/handlers/auth.go",
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
	}
	for _, rel := range files {
		body, err := os.ReadFile(filepath.Join(appDir, rel))
		if err != nil {
			t.Errorf("%s: %v", rel, err)
			continue
		}
		assertGoimportsLocalPrefixOrder(t, rel, string(body), module)
	}
}

func assertGoimportsLocalPrefixOrder(t *testing.T, rel, src, module string) {
	t.Helper()
	groups := parseImportGroups(src)
	if len(groups) == 0 {
		t.Errorf("%s: no import groups", rel)
		return
	}

	rank := map[string]int{"std": 0, "third": 1, "local": 2}
	prev := -1
	seen := map[string]bool{}
	var lastKind string
	for i, group := range groups {
		kind := importKind(group[0], module)
		for _, path := range group {
			if importKind(path, module) != kind {
				t.Errorf("%s: mixed imports in group %d (local must be its own last group): %v", rel, i, group)
				break
			}
		}
		if seen[kind] {
			t.Errorf("%s: duplicate %s import group (goimports wants one group per kind): %v", rel, kind, group)
		}
		seen[kind] = true
		r := rank[kind]
		if r < prev {
			t.Errorf("%s: import group %q after %q; want std → third-party → local", rel, kind, lastKind)
		}
		prev = r
		lastKind = kind
	}
	if seen["local"] && lastKind != "local" {
		t.Errorf("%s: local imports must be the last group, last was %s", rel, lastKind)
	}
	if seen["third"] && seen["local"] && lastKind != "local" {
		t.Errorf("%s: third-party imports must not appear after local", rel)
	}
}

func parseImportGroups(src string) [][]string {
	idx := strings.Index(src, "import (")
	if idx < 0 {
		return nil
	}
	rest := src[idx:]
	end := strings.Index(rest, "\n)")
	if end < 0 {
		return nil
	}
	block := rest[:end]
	var groups [][]string
	var cur []string
	for _, line := range strings.Split(block, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || trim == "import (" || strings.HasPrefix(trim, "//") {
			if trim == "" && len(cur) > 0 {
				groups = append(groups, cur)
				cur = nil
			}
			continue
		}
		path := lastQuoted(trim)
		if path == "" {
			continue
		}
		cur = append(cur, path)
	}
	if len(cur) > 0 {
		groups = append(groups, cur)
	}
	return groups
}

func lastQuoted(line string) string {
	end := strings.LastIndex(line, `"`)
	if end <= 0 {
		return ""
	}
	start := strings.LastIndex(line[:end], `"`)
	if start < 0 {
		return ""
	}
	return line[start+1 : end]
}

func sourceFunc(src, name string) string {
	sig := "func " + name + "("
	idx := strings.Index(src, sig)
	if idx < 0 {
		return ""
	}
	rest := src[idx:]
	next := strings.Index(rest[len(sig):], "\nfunc ")
	if next < 0 {
		return rest
	}
	return rest[:len(sig)+next]
}

func importKind(path, module string) string {
	if path == module || strings.HasPrefix(path, module+"/") {
		return "local"
	}
	first := path
	if i := strings.Index(path, "/"); i >= 0 {
		first = path[:i]
	}
	if strings.Contains(first, ".") {
		return "third"
	}
	return "std"
}
