package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// scaffoldAuth adds login/signup/password-reset to an existing app (cais g auth).
// Templates live in tpl_scaffold_auth.go and tpl_scaffold_auth_pages.go; this file only orchestrates writes and patches.
func scaffoldAuth(dir string, data scaffoldData, dryRun bool) error {
	if _, err := os.Stat(filepath.Join(dir, "internal/handlers/auth.go")); err == nil {
		return fmt.Errorf("auth already exists — remove internal/handlers/auth.go first")
	}

	migrationPath, _, err := nextMigrationFile(dir, "auth", dryRun)
	if err != nil {
		return err
	}

	files := map[string]string{
		"internal/models/user.go":                  tplUserModel,
		"internal/handlers/auth.go":                tplAuthHandler,
		"internal/handlers/auth_test.go":           tplAuthTest,
		"internal/store/password_reset.go":         tplStorePasswordReset,
		migrationPath:                              tplMigration002Auth,
		"web/templates/pages/login.html":           tplPageLogin,
		"web/templates/pages/signup.html":          tplPageSignup,
		"web/templates/pages/forgot_password.html": tplPageForgotPassword,
		"web/templates/pages/reset_password.html":  tplPageResetPassword,
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := writeScaffoldTemplate(full, content, data, path, dryRun); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	if err := patchStoreForAuth(dir, dryRun); err != nil {
		return err
	}
	if err := patchAppForAuth(dir, dryRun); err != nil {
		return err
	}
	if err := patchRoutesForAuth(dir, dryRun); err != nil {
		return err
	}

	return nil
}

func patchStoreForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "FindUserByEmail") {
		return nil
	}

	module := readModulePath(dir)
	modelsImport := fmt.Sprintf(`"%s/internal/models"`, module)

	if !strings.Contains(content, modelsImport) {
		content = strings.Replace(content,
			`import (`,
			fmt.Sprintf(`import (
	%s`, modelsImport),
			1,
		)
	}

	if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/session") {
		content = strings.Replace(content,
			`"github.com/puppe1990/cais/pkg/cais/sqllog"`,
			`"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/cais/pkg/cais/sqllog"`,
			1,
		)
	}

	if !strings.Contains(content, `"errors"`) {
		content = strings.Replace(content,
			`import (
	"database/sql"`,
			`import (
	"database/sql"
	"errors"`,
			1,
		)
	}
	if !strings.Contains(content, `"strings"`) {
		content = strings.Replace(content,
			`"path/filepath"`,
			`"path/filepath"
	"strings"`,
			1,
		)
	}
	if !strings.Contains(content, "ErrEmailTaken") {
		content = strings.Replace(content,
			`)

type Store interface {`,
			`)

var ErrEmailTaken = errors.New("email already registered")

type Store interface {`,
			1,
		)
	}

	ifaceMarker := "\n\tClose() error"
	if !strings.Contains(content, ifaceMarker) {
		return fmt.Errorf("could not patch store interface for auth")
	}
	content = strings.Replace(content,
		ifaceMarker,
		"\n\tFindUserByEmail(email string) (models.User, error)\n\tCreateUser(email, passwordHash string) (int64, error)\n\tCreatePasswordResetToken(userID int64) (string, error)\n\tFindPasswordResetUserID(token string) (int64, bool)\n\tResetPasswordWithToken(token, passwordHash string) error\n\tSessions() session.Store"+ifaceMarker,
		1,
	)

	insert := `
func (s *SQLiteStore) FindUserByEmail(email string) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (s *SQLiteStore) CreateUser(email, passwordHash string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, passwordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, ErrEmailTaken
		}
		return 0, fmt.Errorf("create user: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) Sessions() session.Store {
	return session.NewSQLiteStore(s.db.Raw())
}
`
	content = strings.Replace(content, "\nfunc (s *SQLiteStore) Close()", insert+"\nfunc (s *SQLiteStore) Close()", 1)

	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func patchAppForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/app.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	changed := false

	if !strings.Contains(content, "LoadSession") {
		if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/session") {
			content = strings.Replace(content,
				`"github.com/puppe1990/cais/pkg/cais/middleware"`,
				`"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/cais/pkg/cais/session"`,
				1,
			)
		}
		content = strings.Replace(content,
			"r.Use(middleware.CSRF(cfg))\n",
			"r.Use(middleware.CSRF(cfg))\n\tr.Use(middleware.LoadSession(deps.Store.Sessions()))\n\tr.Use(middleware.Flash)\n",
			1,
		)
		changed = true
	} else if !strings.Contains(content, "middleware.Flash") {
		content = strings.Replace(content,
			"r.Use(middleware.LoadSession(deps.Store.Sessions()))\n",
			"r.Use(middleware.LoadSession(deps.Store.Sessions()))\n\tr.Use(middleware.Flash)\n",
			1,
		)
		changed = true
	}

	if !strings.Contains(content, "SecurityHeaders") {
		content = strings.Replace(content,
			"r.Use(middleware.Recover)\n",
			"r.Use(middleware.Recover)\n\tr.Use(middleware.SecurityHeaders(cfg))\n",
			1,
		)
		changed = true
	}

	if !changed {
		return nil
	}

	return updateScaffoldFile(path, []byte(content), "internal/app/app.go", dryRun)
}

func patchRoutesForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "NewAuthHandler") {
		return nil
	}

	if !strings.Contains(content, "github.com/puppe1990/cais/pkg/cais/middleware") {
		content = strings.Replace(content,
			`"github.com/puppe1990/cais/pkg/cais"`,
			`"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"`,
			1,
		)
	}
	if !strings.Contains(content, `"net/http"`) {
		content = strings.Replace(content,
			`import (
	"github.com/puppe1990/cais/pkg/cais"`,
			`import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"`,
			1,
		)
	}

	insert := `	loginLimit := middleware.NewRateLimiter(10, cfg)
	resetLimit := middleware.NewRateLimiter(10, cfg)

	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, deps.Site, deps.Store.Sessions(), cfg, deps.Catalog)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	r.Get("/signup", auth.SignUp)
	r.Post("/signup", loginLimit.Middleware(http.HandlerFunc(auth.SignUpPost)).ServeHTTP)
	r.Get("/forgot-password", auth.ForgotPassword)
	r.Post("/forgot-password", resetLimit.Middleware(http.HandlerFunc(auth.ForgotPasswordPost)).ServeHTTP)
	r.Get("/reset-password", auth.ResetPassword)
	r.Post("/reset-password", resetLimit.Middleware(http.HandlerFunc(auth.ResetPasswordPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)

`
	content = strings.Replace(content, "func registerRoutes", insert+"func registerRoutes", 1)

	content = strings.Replace(content,
		`r.Get("/dashboard", dashboard.ServeHTTP)`,
		`r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))`,
		1,
	)

	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}
