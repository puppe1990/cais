package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_NewMinimalCreatesSlimApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "slim")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "slim",
		ModulePath: "github.com/puppe1990/slim",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/home.go",
		"go.mod",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/store/migrations/001_contacts.sql",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("minimal app should not have %s", path)
		}
	}
}

func TestCLI_NewCreatesApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "myapp")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "myapp",
		ModulePath: "github.com/puppe1990/myapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/i18n/en.go",
		"internal/i18n/pt.go",
		".env.example",
		"internal/handlers/dashboard.go",
		"web/templates/pages/dashboard.html",
		"web/static/manifest.webmanifest",
		"web/static/js/sw.js",
		"web/static/js/cais.js",
		"web/static/img/go-on-cais.jpg",
		"web/static/og.png",
		"web/static/icons/icon.png",
		"web/templates/partials/chat_sse.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	appGo, err := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(appGo), "WriteTimeout:      0,") {
		t.Error("app.go should disable WriteTimeout for SSE streaming")
	}

	layout, err := os.ReadFile(filepath.Join(appDir, "web/templates/layouts/base.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(layout), "flashMessage") {
		t.Error("base.html should use flashMessage helper")
	}

	css, err := os.ReadFile(filepath.Join(appDir, "input.css"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(css), "fonts.googleapis.com") {
		t.Error("input.css should not import Google Fonts (CSP blocked)")
	}
}

func TestScaffoldNewApp_i18nIncludesSignupKeys(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "i18napp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "i18napp",
		ModulePath: "github.com/puppe1990/i18napp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	en, err := os.ReadFile(filepath.Join(appDir, "internal/i18n/en.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		`"auth.signup_prompt"`,
		`"auth.signup_title"`,
		`"auth.signup_submit"`,
		`"auth.login_prompt"`,
		`"auth.email_taken"`,
	} {
		if !strings.Contains(string(en), key) {
			t.Errorf("internal/i18n/en.go missing %s", key)
		}
	}
}

func TestScaffold_InputCSSIncludesHTMXStyles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "styles")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "styles",
		ModulePath: "github.com/puppe1990/styles",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	css, err := os.ReadFile(filepath.Join(appDir, "input.css"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(css)
	for _, needle := range []string{".htmx-swapping", ".htmx-settling", ".htmx-indicator", ".no-scrollbar", ".cais-toast-enter", ".cais-skeleton"} {
		if !strings.Contains(body, needle) {
			t.Errorf("input.css missing %q", needle)
		}
	}
	tailwind, err := os.ReadFile(filepath.Join(appDir, "tailwind.config.js"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(tailwind), "fonts.googleapis.com") {
		t.Error("tailwind.config.js should not reference Google Fonts")
	}
	if !strings.Contains(string(tailwind), "system-ui") {
		t.Error("tailwind.config.js should use system font stack")
	}
}

func TestScaffold_IncludesQualityTooling(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")

	for _, tc := range []struct {
		name           string
		minimal, blank bool
	}{
		{"full", false, false},
		{"minimal", true, false},
		{"blank", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appDir := filepath.Join(t.TempDir(), tc.name)
			if err := scaffoldNewApp(appDir, scaffoldData{
				AppName:    tc.name,
				ModulePath: "github.com/puppe1990/" + tc.name,
			}, tc.minimal, tc.blank); err != nil {
				t.Fatal(err)
			}

			for _, path := range []string{
				".github/workflows/ci.yml",
				".pre-commit-config.yaml",
				".golangci.yml",
				".prettierrc.json",
				".prettierignore",
			} {
				if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
					t.Errorf("missing %s: %v", path, err)
				}
			}

			makefile, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
			if err != nil {
				t.Fatal(err)
			}
			body := string(makefile)
			for _, target := range []string{"lint:", "format-check:", "pre-commit-install:", "ci:"} {
				if !strings.Contains(body, target) {
					t.Errorf("Makefile missing target %s", target)
				}
			}

			golangci, err := os.ReadFile(filepath.Join(appDir, ".golangci.yml"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(golangci), "github.com/puppe1990/"+tc.name) {
				t.Error(".golangci.yml missing module local-prefix")
			}

			ci, err := os.ReadFile(filepath.Join(appDir, ".github/workflows/ci.yml"))
			if err != nil {
				t.Fatal(err)
			}
			ciBody := string(ci)
			for _, needle := range []string{"go test", "golangci-lint", "prettier", "npm test"} {
				if !strings.Contains(ciBody, needle) {
					t.Errorf("ci.yml missing %q", needle)
				}
			}

			pkg, err := os.ReadFile(filepath.Join(appDir, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(pkg), `"test"`) {
				t.Error("package.json missing test script")
			}
		})
	}
}

func TestScaffoldNewApp_ContactHandlerValidatesName(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "contactapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "contactapp",
		ModulePath: "github.com/puppe1990/contactapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/contact.go"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, `errs.Add("name"`) {
		t.Errorf("contact handler missing name validation: %s", s)
	}
	if !strings.Contains(s, `contact.name_required`) {
		t.Errorf("contact handler missing name_required i18n key: %s", s)
	}
}

func TestScaffoldBlankApp_IncludesSecurityMiddleware(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "blankapp")
	if err := scaffoldNewApp(appDir, scaffoldData{AppName: "blankapp", ModulePath: "github.com/puppe1990/blankapp"}, false, true); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	s := string(body)
	for _, want := range []string{
		"middleware.Recover",
		"middleware.SecurityHeaders(cfg)",
		"ReadHeaderTimeout",
		"ReadTimeout",
		"r.Static",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("blank app missing %q in app.go", want)
		}
	}
}

func TestScaffoldBlankApp_IncludesSessionMiddleware(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "blankapp")
	if err := scaffoldNewApp(appDir, scaffoldData{AppName: "blankapp", ModulePath: "github.com/puppe1990/blankapp"}, false, true); err != nil {
		t.Fatal(err)
	}
	appGo, _ := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	s := string(appGo)
	for _, want := range []string{
		"middleware.LoadSession(deps.Store.Sessions())",
		"r.Use(middleware.Flash)",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("blank app missing %q in app.go", want)
		}
	}
	storeGo, _ := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if !strings.Contains(string(storeGo), "Sessions() session.Store") {
		t.Error("blank store missing Sessions() on interface")
	}
}

func TestCLI_NewBlankCreatesEmptyApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "empty")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "empty",
		ModulePath: "github.com/puppe1990/empty",
	}, false, true); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/app/app.go",
		"internal/app/routes.go",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/home.go",
		"web/templates/pages/home.html",
		"web/templates/layouts/welcome.html",
		"web/templates/partials/cais_logo.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("blank app missing welcome screen file %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/models/contact.go",
		"internal/store/migrations/001_contacts.sql",
		"web/templates/pages/contact.html",
		"web/templates/pages/dashboard.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("blank app should not have %s", path)
		}
	}

	routesBody, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routesBody), "home.ServeHTTP") {
		t.Error("blank app routes should register welcome home handler")
	}
}

func TestScaffoldNewApp_CustomModule(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "myapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "myapp",
		ModulePath: "github.com/acme/myapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/acme/myapp") {
		t.Errorf("go.mod missing custom module path: %s", body)
	}
}

func TestCLI_New_CustomModule(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	appDir := filepath.Join(root, "myapp")

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"new", "myapp", appDir, "--module", "github.com/acme/myapp"}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/acme/myapp") {
		t.Errorf("go.mod missing custom module path: %s", body)
	}
}

func TestCLI_New_CustomModule_DefaultWhenOmitted(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	appDir := filepath.Join(root, "cool-app")

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"new", "cool-app", appDir}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/puppe1990/coolapp") {
		t.Errorf("go.mod missing default module path: %s", body)
	}
}

func TestCLI_New_ModuleRequiresValue(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"new", "myapp", "--module"}); err == nil {
		t.Fatal("expected error for --module without value")
	}
}

func TestCLI_NewMainUsesTemplateHotReload(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "hotreload")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "hotreload",
		ModulePath: "github.com/puppe1990/hotreload",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	mainGo, err := os.ReadFile(filepath.Join(appDir, "cmd/server/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(mainGo)
	if !strings.Contains(body, "NewRendererForEnv") {
		t.Error("main.go should use NewRendererForEnv for development template hot reload")
	}
	air, err := os.ReadFile(filepath.Join(appDir, ".air.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(air), `"html"`) {
		t.Error(".air.toml should not rebuild on html; templates reload from disk in development")
	}
}

func TestCLI_NewIncludesHTMXAndAir(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "full")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "full",
		ModulePath: "github.com/puppe1990/full",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"web/static/js/htmx.min.js",
		"web/static/js/sse-ext.min.js",
		".air.toml",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}
}
