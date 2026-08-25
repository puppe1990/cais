package cli

import (
	"fmt"
	"go/ast"
	"go/token"
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
	content, err := unpatchResourceRoutes(string(body), data)
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

// unpatchResourceRoutes removes only the statements the generator added,
// matched via go/ast (#168). User comments and custom routes that merely
// mention /admin/<plural> survive; braces inside string literals can't corrupt
// block detection because the r.Group statement is removed as one AST node.
func unpatchResourceRoutes(content string, data scaffoldData) (string, error) {
	adminVar := "admin" + data.PluralPascal
	pubVar := lowerFirst(data.PluralPascal)
	adminPaths := map[string]bool{
		"/admin/" + data.Plural:                  true,
		"/admin/" + data.Plural + "/{id}":        true,
		"/admin/" + data.Plural + "/new":         true,
		"/admin/" + data.Plural + "/{id}/edit":   true,
		"/admin/" + data.Plural + "/{id}/delete": true,
	}
	pubPaths := map[string]bool{
		"/" + data.Plural:                  true,
		"/" + data.Plural + "/{id}/toggle": true,
	}
	drop := func(st ast.Stmt) bool {
		return isGeneratedRouteStmt(st, pubVar, adminVar, pubPaths, adminPaths)
	}
	return removeStmtsFromFunc(content, "registerRoutes", drop)
}

func unpatchRoutesForHandler(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Statement-exact removal (#168): the handler's var assignment and its
	// r.Get/r.Post on "/<snake>" calling <camel>.ServeHTTP.
	camel := data.Camel
	drop := func(st ast.Stmt) bool {
		switch s := st.(type) {
		case *ast.AssignStmt:
			for _, lhs := range s.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == camel {
					return true
				}
			}
		case *ast.ExprStmt:
			call, ok := s.X.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return false
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return false
			}
			if recv, ok := sel.X.(*ast.Ident); !ok || recv.Name != "r" {
				return false
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || strings.Trim(lit.Value, "`\"") != "/"+data.Snake {
				return false
			}
			matches := false
			ast.Inspect(call, func(n ast.Node) bool {
				if inner, ok := n.(*ast.SelectorExpr); ok {
					if id, ok := inner.X.(*ast.Ident); ok && id.Name == camel {
						matches = true
					}
				}
				return true
			})
			return matches
		}
		return false
	}
	content, err := removeStmtsFromFunc(string(body), "registerRoutes", drop)
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

func unpatchStoreForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := removeStoreResourceMethods(string(body), data)
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

// removeStoreResourceMethods removes generated methods and interface entries by
// exact name (#168): a hand-written countBookmarksByUser survives destroying
// resource bookmark even though countBookmarks does not.
func removeStoreResourceMethods(content string, data scaffoldData) (string, error) {
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
	names := make(map[string]bool, len(patterns))
	for _, p := range patterns {
		names[p] = true
	}
	content, err := removeDeclsByName(content, names)
	if err != nil {
		return "", err
	}
	content, err = removeInterfaceMethods(content, "Store", names)
	if err != nil {
		return "", err
	}
	content = cleanupStoreImports(content)
	content = regexp.MustCompile(`\nfunc strPtr\(s string\) \*string \{ return &s \}\n`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`\nfunc int64Ptr\(n int64\) \*int64 \{ return &n \}\n`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`\nfunc boolInt\(v bool\) int \{[^}]+\}\n`).ReplaceAllString(content, "\n")
	return content, nil
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

func unpatchStoreTestForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content, err := removeDeclsByName(string(body), map[string]bool{"TestStore_Insert" + data.Pascal: true})
	if err != nil {
		return err
	}
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
