package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/puppe1990/cais/pkg/cais/pwa"
)

// qualityToolingFiles returns CI/lint/format configs written by cais new and cais g ci.
// Kept in a helper so minimal/blank/full scaffolds share the same tooling block.
func qualityToolingFiles() map[string]string {
	return map[string]string{
		".github/workflows/ci.yml": tplCIWorkflow,
		".pre-commit-config.yaml":  tplPreCommitConfig,
		".golangci.yml":            tplGolangci,
		".prettierrc.json":         tplPrettierrc,
		".prettierignore":          tplPrettierignore,
	}
}

func scaffoldNewApp(dir string, data scaffoldData, minimal bool, blank bool) error {
	files := map[string]string{
		"go.mod":                                     tplGoMod,
		"cmd/server/main.go":                         tplMain,
		"cmd/console/main.go":                        tplConsole,
		"internal/app/app.go":                        tplApp,
		"internal/app/routes.go":                     tplRoutes,
		"internal/handlers/home.go":                  tplHomeHandler,
		"internal/handlers/home_test.go":             tplHomeTest,
		"internal/handlers/contact.go":               tplContactHandler,
		"internal/handlers/contact_test.go":          tplContactTest,
		"internal/handlers/dashboard.go":             tplDashboardHandler,
		"internal/handlers/dashboard_test.go":        tplDashboardTest,
		"internal/handlers/inertia_test.go":          tplInertiaTest,
		"internal/handlers/helpers_test.go":          tplHelpersTest,
		"internal/models/contact.go":                 tplContactModel,
		"internal/store/store.go":                    tplStore,
		"internal/store/password_reset.go":           tplStorePasswordReset,
		"internal/store/store_test.go":               tplStoreTest,
		"internal/store/migrations.go":               tplMigrations,
		"internal/store/migrations/001_contacts.sql": tplMigration001,
		"internal/store/migrations/002_auth.sql":     tplMigration002Auth,
		"internal/models/user.go":                    tplUserModel,
		"internal/handlers/auth.go":                  tplAuthHandler,
		"internal/handlers/auth_test.go":             tplAuthTest,
		"internal/handlers/auth_signup_test.go":      tplAuthSignupTest,
		"internal/handlers/auth_reset_test.go":       tplAuthResetTest,
		"web/embed.go":                               tplWebEmbed,
		"web/templates/app.html":                     tplAppHTML,
		"web/src/main.js":                            tplMainJS,
		"web/src/components/AppLayout.svelte":        tplAppLayout,
		"web/src/pages/Home.svelte":                  tplSvelteHome,
		"web/src/pages/Contact.svelte":               tplSvelteContact,
		"web/src/pages/Dashboard.svelte":             tplSvelteDashboard,
		"web/src/pages/Login.svelte":                 tplSvelteLogin,
		"web/src/pages/Signup.svelte":                tplSvelteSignup,
		"web/src/pages/ForgotPassword.svelte":        tplSvelteForgotPassword,
		"web/src/pages/ResetPassword.svelte":         tplSvelteResetPassword,
		"web/static/build/.gitkeep":                  tplBuildGitkeep,
		"web/static/css/styles.css":                  tplEmptyCSS,
		"input.css":                                  tplInputCSS,
		"tailwind.config.js":                         tplTailwind,
		"vite.config.js":                             tplViteConfig,
		"svelte.config.js":                           tplSvelteConfig,
		"vitest-setup.js":                            tplVitestSetup,
		"package.json":                               tplPackageJSON,
		"Makefile":                                   tplMakefile,
		".gitignore":                                 tplGitignore,
		".air.toml":                                  tplAir,
		".env.example":                               tplEnvExample,
		"README.md":                                  tplREADME,
		"internal/i18n/i18n.go":                      tplI18nCatalog,
		"internal/i18n/en.go":                        tplI18nEn,
		"internal/i18n/pt.go":                        tplI18nPt,
		"internal/i18n/i18n_test.go":                 tplI18nTest,
		"internal/db/seeds.go":                       tplSeeds,
	}
	for path, content := range qualityToolingFiles() {
		files[path] = content
	}

	if blank {
		files = map[string]string{
			"go.mod":                             tplGoMod,
			"cmd/server/main.go":                 tplMainBlank,
			"cmd/console/main.go":                tplConsole,
			"internal/app/app.go":                tplAppBlank,
			"internal/app/routes.go":             tplRoutesBlank,
			"internal/handlers/helpers_test.go":  tplHelpersTest,
			"internal/handlers/inertia_test.go":  tplInertiaTest,
			"internal/store/store.go":            tplStoreMinimal,
			"internal/store/store_test.go":       tplStoreTestMinimal,
			"internal/store/migrations.go":       tplMigrations,
			"internal/store/migrations/.gitkeep": "",
			"web/embed.go":                       tplWebEmbed,
			"internal/handlers/home.go":          tplHomeHandler,
			"internal/handlers/home_test.go":     tplHomeTest,
			"web/templates/app.html":             tplAppHTML,
			"web/src/main.js":                    tplMainJS,
			"web/src/components/AppLayout.svelte": tplAppLayout,
			"web/src/pages/Home.svelte":          tplSvelteHome,
			"web/static/build/.gitkeep":          tplBuildGitkeep,
			"web/static/css/styles.css":          tplEmptyCSS,
			"input.css":                          tplInputCSS,
			"tailwind.config.js":                 tplTailwind,
			"vite.config.js":                     tplViteConfig,
			"svelte.config.js":                   tplSvelteConfig,
			"vitest-setup.js":                    tplVitestSetup,
			"package.json":                       tplPackageJSON,
			"Makefile":                           tplMakefile,
			".gitignore":                         tplGitignore,
			".air.toml":                          tplAir,
			".env.example":                       tplEnvExample,
			"README.md":                          tplREADMEBlank,
			"internal/i18n/i18n.go":              tplI18nCatalog,
			"internal/i18n/en.go":                tplI18nEn,
			"internal/i18n/pt.go":                tplI18nPt,
			"internal/i18n/i18n_test.go":         tplI18nTest,
			"internal/db/seeds.go":               tplSeedsMinimal,
		}
		for path, content := range qualityToolingFiles() {
			files[path] = content
		}
	} else if minimal {
		delete(files, "internal/handlers/contact.go")
		delete(files, "internal/handlers/contact_test.go")
		delete(files, "internal/handlers/dashboard.go")
		delete(files, "internal/handlers/dashboard_test.go")
		delete(files, "internal/handlers/auth.go")
		delete(files, "internal/handlers/auth_test.go")
		delete(files, "internal/handlers/auth_signup_test.go")
		delete(files, "internal/handlers/auth_reset_test.go")
		delete(files, "internal/models/contact.go")
		delete(files, "internal/models/user.go")
		delete(files, "internal/store/migrations/001_contacts.sql")
		delete(files, "internal/store/migrations/002_auth.sql")
		delete(files, "web/src/pages/Contact.svelte")
		delete(files, "web/src/pages/Dashboard.svelte")
		delete(files, "web/src/pages/Login.svelte")
		delete(files, "web/src/pages/Signup.svelte")
		delete(files, "web/src/pages/ForgotPassword.svelte")
		delete(files, "web/src/pages/ResetPassword.svelte")
		files["internal/app/routes.go"] = tplRoutesMinimal
		files["internal/store/store.go"] = tplStoreMinimal
		files["internal/store/store_test.go"] = tplStoreTestMinimal
		files["internal/handlers/home_test.go"] = tplHomeTest
		files["internal/store/migrations/.gitkeep"] = ""
		files["internal/db/seeds.go"] = tplSeedsMinimal
	}

	for path, content := range files {
		if err := writeTemplate(filepath.Join(dir, path), content, data); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	if err := gofmtGoFiles(dir); err != nil {
		return fmt.Errorf("gofmt: %w", err)
	}

	if err := pwa.InstallForInertia(dir, data.AppName); err != nil {
		return fmt.Errorf("pwa assets: %w", err)
	}

	if err := patchGoModReplace(dir); err != nil {
		return err
	}

	if os.Getenv("CAIS_SKIP_TIDY") == "1" {
		return nil
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func scaffoldHandler(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	files := map[string]string{
		filepath.Join("internal/handlers", data.Snake+".go"):      tplGenericHandler,
		filepath.Join("internal/handlers", data.Snake+"_test.go"): tplGenericHandlerTest,
		filepath.Join("web/src/pages", data.Pascal+".svelte"):     tplGenericPage,
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeScaffoldTemplate(full, content, data, path, dryRun); err != nil {
			return err
		}
	}

	return patchRoutes(dir, data, dryRun)
}

func scaffoldPage(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	rel := filepath.Join("web/src/pages", data.Pascal+".svelte")
	path := filepath.Join(dir, rel)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("web/src/pages/%s.svelte already exists", data.Pascal)
	}
	return writeScaffoldTemplate(path, tplGenericPage, data, rel, dryRun)
}

func scaffoldMigration(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	rel, _, err := nextMigrationFile(dir, data.Snake, dryRun)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, rel)
	content := fmt.Sprintf("-- migration: %s\n-- up\n\n-- down\n\n", data.Snake)
	return writeScaffoldFile(path, []byte(content), 0o644, rel, dryRun)
}

func patchRoutes(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if strings.Contains(string(body), "New"+data.Pascal+"Handler") {
		return nil
	}

	insert := fmt.Sprintf(
		"\n\t%s := handlers.New%sHandler(deps.Site, deps.Catalog, deps.Inertia)\n\tr.Get(\"/%s\", %s.ServeHTTP)\n",
		data.Camel, data.Pascal, data.Snake, data.Camel,
	)

	content := string(body)
	updated, err := insertBeforeFunctionEnd(content, "registerRoutes", insert)
	if err != nil {
		return fmt.Errorf("could not patch routes.go: %w", err)
	}
	return updateScaffoldFile(path, []byte(updated), "internal/app/routes.go", dryRun)
}

func writeTemplate(path, tpl string, data scaffoldData) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if tpl == "" {
		return os.WriteFile(path, nil, 0o644)
	}
	t, err := template.New("scaffold").Parse(tpl)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return t.Execute(f, data)
}

func gofmtGoFiles(dir string) error {
	cmd := exec.Command("gofmt", "-w", "./internal/", "./cmd/")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
