package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func destroyResource(dir, name string, dryRun bool) error {
	data := dataForResource(name)

	files := []string{
		filepath.Join("internal/models", data.Snake+".go"),
		filepath.Join("internal/handlers", "admin_"+data.Plural+".go"),
		filepath.Join("internal/handlers", "admin_"+data.Plural+"_test.go"),
		filepath.Join("web/templates/pages", "admin_"+data.Plural+".html"),
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_show.html"),
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_form.html"),
		filepath.Join("web/templates/partials", "admin_"+data.Snake+"_form_errors.html"),
		filepath.Join("web/templates/partials", "admin_"+data.Plural+"_index.html"),
		filepath.Join("web/src/pages", "Admin"+data.PluralPascal+".svelte"),
		filepath.Join("web/src/pages", "Admin"+data.Pascal+"Form.svelte"),
		filepath.Join("web/src/pages", "Admin"+data.Pascal+"Show.svelte"),
		filepath.Join("internal/handlers", data.Plural+".go"),
		filepath.Join("internal/handlers", data.Plural+"_test.go"),
		filepath.Join("web/templates/pages", data.Plural+".html"),
		filepath.Join("web/templates/partials", data.Plural+"_toggle.html"),
		filepath.Join("web/templates/partials", data.Plural+"_list.html"),
		filepath.Join("web/src/pages", data.PluralPascal+".svelte"),
	}

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, _ := os.ReadDir(migrationsDir)
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_"+data.Plural+".sql") {
			files = append(files, filepath.Join("internal/store/migrations", e.Name()))
		}
	}

	for _, rel := range files {
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		if dryRun {
			printfScaffold("remove", rel)
			continue
		}
		if err := os.Remove(full); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
	}

	if err := unpatchRoutesForResource(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchStoreForResource(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchStoreTestForResource(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchSeedsForResource(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchMainForSeed(dir, data, dryRun); err != nil {
		return err
	}
	return unpatchLayoutNavForResource(dir, data, dryRun)
}

func destroyModel(dir, name string, dryRun bool) error {
	data := dataForResource(name)

	files := []string{
		filepath.Join("internal/models", data.Snake+".go"),
	}

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, _ := os.ReadDir(migrationsDir)
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_"+data.Plural+".sql") {
			files = append(files, filepath.Join("internal/store/migrations", e.Name()))
		}
	}

	for _, rel := range files {
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		if dryRun {
			printfScaffold("remove", rel)
			continue
		}
		if err := os.Remove(full); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
	}

	return unpatchStoreForResource(dir, data, dryRun)
}

func destroyHandler(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	files := []string{
		filepath.Join("internal/handlers", data.Snake+".go"),
		filepath.Join("internal/handlers", data.Snake+"_test.go"),
		filepath.Join("web/templates/pages", data.Snake+".html"),
	}
	for _, rel := range files {
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		if dryRun {
			printfScaffold("remove", rel)
			continue
		}
		if err := os.Remove(full); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
	}
	return unpatchRoutesForHandler(dir, data, dryRun)
}

func unpatchRoutesForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := unpatchResourceRoutes(string(body), data)
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

func unpatchResourceRoutes(content string, data scaffoldData) string {
	lines := strings.Split(content, "\n")
	var out []string
	skipUntil := -1
	adminVar := "admin" + data.PluralPascal
	pubVar := lowerFirst(data.PluralPascal)

	for i, line := range lines {
		if i <= skipUntil {
			continue
		}

		if strings.Contains(line, adminVar+" :=") || strings.Contains(line, "NewAdmin"+data.PluralPascal+"Handler") {
			if j, ok := nextRouteGroupBlock(lines, i+1); ok {
				skipUntil = blockEndAt(lines, j)
			}
			continue
		}

		if strings.Contains(line, pubVar+" :=") || strings.Contains(line, "New"+data.PluralPascal+"Handler(") {
			continue
		}
		if strings.Contains(line, `"/`+data.Plural+`"`) && strings.Contains(line, pubVar) {
			continue
		}

		if strings.Contains(line, `"/admin/`+data.Plural) {
			if strings.Contains(strings.TrimSpace(line), "r.Group(") {
				skipUntil = blockEndAt(lines, i)
			}
			continue
		}
		if strings.Contains(line, adminVar+".") {
			continue
		}

		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func nextRouteGroupBlock(lines []string, from int) (int, bool) {
	limit := from + 3
	if limit > len(lines) {
		limit = len(lines)
	}
	for j := from; j < limit; j++ {
		if strings.Contains(lines[j], "r.Group(") {
			return j, true
		}
		trim := strings.TrimSpace(lines[j])
		if trim != "" && !strings.HasPrefix(trim, "//") {
			break
		}
	}
	return 0, false
}

func blockEndAt(lines []string, start int) int {
	depth := 0
	for j := start; j < len(lines); j++ {
		depth += strings.Count(lines[j], "{") - strings.Count(lines[j], "}")
		if depth <= 0 {
			return j
		}
	}
	return len(lines) - 1
}

func unpatchRoutesForHandler(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(body), "\n")
	var out []string
	for _, line := range lines {
		if strings.Contains(line, "New"+data.Pascal+"Handler") ||
			strings.Contains(line, `"/`+data.Snake+`"`) && strings.Contains(line, data.Camel+".ServeHTTP") {
			continue
		}
		out = append(out, line)
	}
	return updateScaffoldFile(path, []byte(strings.Join(out, "\n")), "internal/app/routes.go", dryRun)
}

func unpatchStoreForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := removeStoreResourceMethods(string(body), data)
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func removeStoreResourceMethods(content string, data scaffoldData) string {
	patterns := []string{
		"Insert" + data.Pascal,
		"Update" + data.Pascal,
		"Delete" + data.Pascal,
		"Find" + data.Pascal + "ByID",
		"ListAll" + data.PluralPascal,
		"List" + data.PluralPascal,
		"SeedDemo" + data.PluralPascal,
		"count" + data.PluralPascal,
	}
	for _, p := range patterns {
		content = removeGoFunc(content, p)
	}
	content = removeMethodsFromStoreInterface(content, patterns)
	content = cleanupStoreImports(content)
	content = regexp.MustCompile(`\nfunc strPtr\(s string\) \*string \{ return &s \}\n`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`\nfunc int64Ptr\(n int64\) \*int64 \{ return &n \}\n`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`\nfunc boolInt\(v bool\) int \{[^}]+\}\n`).ReplaceAllString(content, "\n")
	return content
}

func removeMethodsFromStoreInterface(content string, methodPrefixes []string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inStoreIface := false
	depth := 0

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if !inStoreIface && strings.HasPrefix(trim, "type Store interface") {
			inStoreIface = true
			depth = strings.Count(line, "{") - strings.Count(line, "}")
			out = append(out, line)
			continue
		}
		if inStoreIface {
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			if storeInterfaceMethodLine(line, methodPrefixes) {
				if depth <= 0 {
					inStoreIface = false
				}
				continue
			}
			out = append(out, line)
			if depth <= 0 {
				inStoreIface = false
			}
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func storeInterfaceMethodLine(line string, methodPrefixes []string) bool {
	if !strings.HasPrefix(line, "\t") || strings.Contains(line, "type ") {
		return false
	}
	for _, prefix := range methodPrefixes {
		if strings.Contains(line, prefix+"(") {
			return true
		}
	}
	return false
}

func cleanupStoreImports(content string) string {
	if !strings.Contains(content, "models.") {
		re := regexp.MustCompile(`(?m)^\s*"[^"]+/internal/models"\s*\n`)
		content = re.ReplaceAllString(content, "")
		content = regexp.MustCompile(`import \(\n\n`).ReplaceAllString(content, "import (\n")
	}
	if !strings.Contains(content, "pagination.") {
		content = strings.Replace(content, "\t\""+frameworkModule+"/pkg/cais/pagination\"\n", "", 1)
	}
	return content
}

func removeGoFunc(content, namePrefix string) string {
	return removeFuncBlock(content, "func (s *SQLiteStore) "+namePrefix)
}

func removeFuncBlock(content, signaturePrefix string) string {
	lines := strings.Split(content, "\n")
	var out []string
	skipping := false
	depth := 0
	for _, line := range lines {
		if !skipping && strings.Contains(line, signaturePrefix) {
			skipping = true
			depth = strings.Count(line, "{") - strings.Count(line, "}")
			if depth <= 0 {
				depth = 1
			}
			continue
		}
		if skipping {
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			if depth <= 0 {
				skipping = false
			}
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func unpatchStoreTestForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := removeGoFunc(string(body), "TestStore_Insert"+data.Pascal)
	return updateScaffoldFile(path, []byte(content), "internal/store/store_test.go", dryRun)
}

func unpatchSeedsForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	block := fmt.Sprintf("\tif err := s.SeedDemo%s(); err != nil {\n\t\treturn err\n\t}\n", data.PluralPascal)
	content := strings.Replace(string(body), block, "", 1)
	return updateScaffoldFile(path, []byte(content), "internal/db/seeds.go", dryRun)
}

func unpatchMainForSeed(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "cmd/server/main.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	block := fmt.Sprintf(`
	if err := s.SeedDemo%s(); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("seed: %%w", err)
	}
`, data.PluralPascal)
	content := strings.Replace(string(body), block, "", 1)
	return updateScaffoldFile(path, []byte(content), "cmd/server/main.go", dryRun)
}

func unpatchLayoutNavForResource(dir string, data scaffoldData, dryRun bool) error {
	path, inertia := layoutNavFile(dir)
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	link := publicNavLink(data, inertia)
	content := strings.Replace(string(body), link, "", 1)
	rel := strings.TrimPrefix(path, dir+string(os.PathSeparator))
	return updateScaffoldFile(path, []byte(content), rel, dryRun)
}

func (c *CLI) cmdDestroy(args []string) error {
	dryRun := false
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			continue
		}
		filtered = append(filtered, arg)
	}
	args = filtered
	setScaffoldOut(c.Out)

	if len(args) < 1 {
		return fmt.Errorf("usage: cais destroy [--dry-run] <resource|handler|model|auth|migration> [name]")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if !isCaisApp(cwd) {
		return fmt.Errorf("not a Cais app")
	}

	kind := args[0]
	var genErr error
	switch kind {
	case "resource", "handler", "model", "migration":
		if len(args) < 2 {
			return fmt.Errorf("usage: cais destroy [--dry-run] %s <name>", kind)
		}
		name := args[1]
		switch kind {
		case "resource":
			genErr = destroyResource(cwd, name, dryRun)
		case "handler":
			genErr = destroyHandler(cwd, name, dryRun)
		case "model":
			genErr = destroyModel(cwd, name, dryRun)
		case "migration":
			genErr = destroyMigration(cwd, name, dryRun)
		}
		if genErr != nil {
			return genErr
		}
		if !dryRun {
			_, _ = fmt.Fprintf(c.Out, "=> Removed %s %s\n", kind, name)
		}
	case "auth":
		if len(args) > 1 {
			return fmt.Errorf("usage: cais destroy [--dry-run] auth")
		}
		genErr = destroyAuth(cwd, dryRun)
		if genErr != nil {
			return genErr
		}
		if !dryRun {
			_, _ = fmt.Fprintln(c.Out, "=> Removed auth")
		}
	default:
		return fmt.Errorf("unknown destroy target %q (use resource, handler, model, auth, or migration)", kind)
	}
	return nil
}
