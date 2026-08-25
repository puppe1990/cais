package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func destroyAuth(dir string, dryRun bool) error {
	files := []string{
		"internal/models/user.go",
		"internal/handlers/auth.go",
		"internal/handlers/auth_test.go",
		"web/src/pages/Login.svelte",
		"web/src/pages/Signup.svelte",
		"web/src/pages/ForgotPassword.svelte",
		"web/src/pages/ResetPassword.svelte",
		"web/src/components/AuthLayout.svelte",
		// Legacy HTMX pages (pre-Inertia auth generator)
		"web/templates/pages/login.html",
		"web/templates/pages/signup.html",
		"web/templates/pages/forgot_password.html",
		"web/templates/pages/reset_password.html",
	}

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, _ := os.ReadDir(migrationsDir)
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_auth.sql") {
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

	if err := unpatchStoreForAuth(dir, dryRun); err != nil {
		return err
	}
	if err := unpatchAppForAuth(dir, dryRun); err != nil {
		return err
	}
	return unpatchRoutesForAuthDestroy(dir, dryRun)
}

func unpatchStoreForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := removeStoreAuthMethods(string(body))
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func removeStoreAuthMethods(content string) (string, error) {
	patterns := []string{
		"FindUserByEmail",
		"CreateUser",
		"CreatePasswordResetToken",
		"FindPasswordResetUserID",
		"ResetPasswordWithToken",
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
	content = cleanupSessionImport(content)
	return content, nil
}

func cleanupSessionImport(content string) string {
	if strings.Contains(content, "session.") {
		return content
	}
	content = strings.Replace(content, "\t\"github.com/puppe1990/cais/pkg/cais/session\"\n", "", 1)
	content = regexp.MustCompile(`import \(\n\n`).ReplaceAllString(content, "import (\n")
	return content
}

func unpatchAppForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/app.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := unpatchAuthMiddleware(string(body))
	return updateScaffoldFile(path, []byte(content), "internal/app/app.go", dryRun)
}

// unpatchAuthMiddleware is a no-op: LoadSession/Flash are baseline middleware in full and blank apps.
func unpatchAuthMiddleware(content string) string {
	return content
}

func unpatchRoutesForAuthDestroy(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := unpatchAuthRoutes(string(body))
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

func unpatchAuthRoutes(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	for _, line := range lines {
		if strings.Contains(line, "loginLimit :=") ||
			strings.Contains(line, "NewAuthHandler") ||
			strings.Contains(line, `"/login"`) ||
			strings.Contains(line, `"/logout"`) ||
			strings.Contains(line, "auth.Login") ||
			strings.Contains(line, "auth.LogoutPost") {
			continue
		}
		if strings.Contains(line, "RequireAuthFunc") && strings.Contains(line, "dashboard") {
			line = strings.Replace(line,
				`middleware.RequireAuthFunc("/login", dashboard.ServeHTTP)`,
				`dashboard.ServeHTTP`,
				1,
			)
		}
		out = append(out, line)
	}
	content = strings.Join(out, "\n")
	if !strings.Contains(content, "middleware.") {
		content = strings.Replace(content,
			`"github.com/puppe1990/cais/pkg/cais/middleware"
`,
			"",
			1,
		)
		content = strings.Replace(content,
			`
	"github.com/puppe1990/cais/pkg/cais/middleware"`,
			"",
			1,
		)
	}
	if !strings.Contains(content, "http.HandlerFunc") && !strings.Contains(content, "http.") {
		lines = strings.Split(content, "\n")
		out = nil
		for _, line := range lines {
			if line == "\t\"net/http\"" || line == `"net/http"` {
				continue
			}
			out = append(out, line)
		}
		content = strings.Join(out, "\n")
	}
	return content
}

func destroyMigration(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	suffix := "_" + data.Snake + ".sql"

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var removed int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		rel := filepath.Join("internal/store/migrations", e.Name())
		if dryRun {
			printfScaffold("remove", rel)
			removed++
			continue
		}
		if err := os.Remove(filepath.Join(dir, rel)); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
		removed++
	}

	if removed == 0 {
		return fmt.Errorf("no migration matching %q", suffix)
	}
	return nil
}
