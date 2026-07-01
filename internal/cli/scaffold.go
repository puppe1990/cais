package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/matheuspuppe/cais/pkg/cais/pwa"
)

func scaffoldNewApp(dir string, data scaffoldData, minimal bool) error {
	files := map[string]string{
		"go.mod":                                      tplGoMod,
		"cmd/server/main.go":                          tplMain,
		"internal/app/app.go":                         tplApp,
		"internal/app/routes.go":                      tplRoutes,
		"internal/handlers/home.go":                   tplHomeHandler,
		"internal/handlers/home_test.go":              tplHomeTest,
		"internal/handlers/contact.go":                tplContactHandler,
		"internal/handlers/contact_test.go":           tplContactTest,
		"internal/handlers/dashboard.go":              tplDashboardHandler,
		"internal/handlers/dashboard_test.go":         tplDashboardTest,
		"internal/handlers/helpers_test.go":           tplHelpersTest,
		"internal/models/contact.go":                  tplContactModel,
		"internal/store/store.go":                     tplStore,
		"internal/store/store_test.go":                tplStoreTest,
		"internal/store/migrations.go":                tplMigrations,
		"internal/store/migrations/001_contacts.sql":  tplMigration001,
		"web/embed.go":                                tplWebEmbed,
		"web/templates/layouts/base.html":             tplLayout,
		"web/templates/pages/home.html":               tplPageHome,
		"web/templates/pages/contact.html":            tplPageContact,
		"web/templates/pages/dashboard.html":          tplPageDashboard,
		"web/templates/partials/contact_errors.html":  tplPartialErrors,
		"web/templates/partials/contact_success.html": tplPartialSuccess,
		"web/static/js/.gitkeep":                      "",
		"input.css":                                   tplInputCSS,
		"tailwind.config.js":                          tplTailwind,
		"package.json":                                tplPackageJSON,
		"Makefile":                                    tplMakefile,
		".gitignore":                                  tplGitignore,
		"README.md":                                   tplREADME,
	}

	if minimal {
		delete(files, "internal/handlers/contact.go")
		delete(files, "internal/handlers/contact_test.go")
		delete(files, "internal/handlers/dashboard.go")
		delete(files, "internal/handlers/dashboard_test.go")
		delete(files, "internal/models/contact.go")
		delete(files, "internal/store/migrations/001_contacts.sql")
		delete(files, "web/templates/pages/contact.html")
		delete(files, "web/templates/pages/dashboard.html")
		delete(files, "web/templates/partials/contact_errors.html")
		delete(files, "web/templates/partials/contact_success.html")
		files["internal/app/routes.go"] = tplRoutesMinimal
		files["internal/store/store.go"] = tplStoreMinimal
		files["internal/store/store_test.go"] = tplStoreTestMinimal
		files["web/templates/layouts/base.html"] = tplLayoutMinimal
		files["internal/store/migrations/.gitkeep"] = ""
	}

	for path, content := range files {
		if err := writeTemplate(filepath.Join(dir, path), content, data); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	if err := pwa.InstallTo(dir, data.AppName); err != nil {
		return fmt.Errorf("pwa assets: %w", err)
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

func scaffoldHandler(dir, name string) error {
	data := dataForHandler(name)
	files := map[string]string{
		filepath.Join("internal/handlers", data.Snake+".go"):      tplGenericHandler,
		filepath.Join("internal/handlers", data.Snake+"_test.go"): tplGenericHandlerTest,
		filepath.Join("web/templates/pages", data.Snake+".html"):  tplGenericPage,
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeTemplate(full, content, data); err != nil {
			return err
		}
		_, _ = fmt.Printf("  create %s\n", path)
	}

	return patchRoutes(dir, data)
}

func scaffoldPage(dir, name string) error {
	data := dataForHandler(name)
	path := filepath.Join(dir, "web/templates/pages", data.Snake+".html")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("web/templates/pages/%s.html already exists", data.Snake)
	}
	if err := writeTemplate(path, tplGenericPage, data); err != nil {
		return err
	}
	_, _ = fmt.Printf("  create web/templates/pages/%s.html\n", data.Snake)
	return nil
}

func scaffoldMigration(dir, name string) error {
	data := dataForHandler(name)
	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	next := len(entries) + 1
	filename := fmt.Sprintf("%03d_%s.sql", next, data.Snake)
	path := filepath.Join(migrationsDir, filename)
	content := fmt.Sprintf("-- migration: %s\n", data.Snake)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Printf("  create internal/store/migrations/%s\n", filename)
	return nil
}

func patchRoutes(dir string, data scaffoldData) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if strings.Contains(string(body), "New"+data.Pascal+"Handler") {
		return nil
	}

	insert := fmt.Sprintf(
		"\t%s := handlers.New%sHandler(deps.Renderer)\n\tr.Get(\"/%s\", %s.ServeHTTP)\n",
		data.Camel, data.Pascal, data.Snake, data.Camel,
	)

	content := string(body)
	marker := "\n}\n"
	idx := strings.LastIndex(content, marker)
	if idx == -1 {
		return fmt.Errorf("could not patch routes.go")
	}

	updated := content[:idx] + insert + content[idx:]
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update internal/app/routes.go")
	return nil
}

func renderSnippet(tpl string, data scaffoldData) (string, error) {
	t, err := template.New("snippet").Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
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
