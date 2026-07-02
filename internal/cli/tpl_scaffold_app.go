package cli

const tplMain = `package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/boot"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"{{.ModulePath}}/internal/app"
	appi18n "{{.ModulePath}}/internal/i18n"
	"{{.ModulePath}}/internal/store"
	"{{.ModulePath}}/web"
)

func main() {
	cfg := cais.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	preferredPort := cfg.Port
	port, shifted, err := cais.ResolvePort(cfg.Port, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Port = port

	a, err := bootstrapWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	shiftedFrom := ""
	if shifted {
		shiftedFrom = preferredPort
	}
	boot.Print(os.Stdout, boot.Options{
		AppName:         "{{.AppName}}",
		Config:          cfg,
		Version:         boot.CaisVersion(),
		PortShiftedFrom: shiftedFrom,
	})
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func bootstrap() (*app.App, error) {
	return bootstrapWithConfig(cais.Load())
}

func bootstrapWithConfig(cfg cais.Config) (*app.App, error) {
	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	catalog := appi18n.NewCatalog(cfg.Locale)
	renderer, err := cais.NewRenderer(tmplFS, catalog)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := findWebDir("static")
	if err != nil {
		_ = s.Close()
		return nil, err
	}

	return app.New(cfg, app.Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: staticDir,
		Site:      meta.SiteFrom("{{.AppName}}", cfg.AppURL),
		Catalog:   catalog,
	})
}

func findWebDir(subpath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(wd, "web", subpath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("web/%s not found", subpath)
		}
		wd = parent
	}
}
`

const tplConsole = `package main

import (
	"log"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/console"
	"{{.ModulePath}}/internal/store"
)

func openStore(cfg cais.Config) (*store.SQLiteStore, error) {
	return store.NewSQLiteStore(cfg.DBPath, cfg.Env)
}

func bindings(s *store.SQLiteStore) map[string]any {
	return map[string]any{
		"store": s,
		"db":    s.DB(),
	}
}

func main() {
	cfg := cais.Load()
	s, err := openStore(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	active := s
	if err := console.Run(console.Options{
		AppName:  "{{.AppName}}",
		Config:   cfg,
		Bindings: bindings(active),
		Reload: func() (map[string]any, error) {
			_ = active.Close()
			next, err := openStore(cfg)
			if err != nil {
				return nil, err
			}
			active = next
			return bindings(active), nil
		},
	}); err != nil {
		log.Fatal(err)
	}
}
`

const tplApp = `package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"{{.ModulePath}}/internal/store"
)

type Deps struct {
	Renderer  *cais.Renderer
	Store     store.Store
	StaticDir string
	Site      meta.Site
	Catalog   *i18n.Catalog
}

type App struct {
	config cais.Config
	store  store.Store
	router *cais.Router
	server *http.Server
}

func New(cfg cais.Config, deps Deps) (*App, error) {
	if deps.Renderer == nil {
		return nil, fmt.Errorf("renderer is required")
	}
	if deps.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	site := deps.Site
	if site.AppName == "" {
		site = meta.SiteFrom("{{.AppName}}", cfg.AppURL)
	}
	deps.Site = site

	r := cais.NewRouter()
	r.Use(middleware.CSRF(cfg))
	r.Use(middleware.LoadSession(deps.Store.Sessions()))
	r.Use(middleware.Flash)
	buf := devlog.Prepare(cfg.Env)
	if buf != nil {
		r.Use(middleware.LoggerTo(cfg, devlog.MirrorDefault(log.Writer())))
	} else {
		r.Use(middleware.Logger(cfg))
	}
	r.Use(middleware.Recover)
	r.Use(middleware.SecurityHeaders(cfg))
	r.Static("/static", deps.StaticDir)

	registerRoutes(r, deps, cfg)
	devlog.Register(r, cfg.Env, buf)
	r.Get("/health", healthHandler(deps.Store))

	return &App{
		config: cfg,
		store:  deps.Store,
		router: r,
		server: &http.Server{
			Addr:              cfg.Port,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

func healthHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if err := s.Ping(); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}

func (a *App) Handler() http.Handler {
	return a.router
}

func (a *App) Run() error {
	return a.RunContext(context.Background())
}

func (a *App) RunContext(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			_ = a.store.Close()
			return err
		}
		<-errCh
		_ = a.store.Close()
		return nil
	case err := <-errCh:
		_ = a.store.Close()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
`

const tplRoutes = `package app

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"{{.ModulePath}}/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, deps.Site, deps.Catalog, cfg)
	dashboard := handlers.NewDashboardHandler(deps.Renderer, deps.Store, deps.Site, cfg)
	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, deps.Site, deps.Store.Sessions(), cfg, deps.Catalog)

	loginLimit := middleware.NewRateLimiter(10, cfg)
	resetLimit := middleware.NewRateLimiter(10, cfg)
	contactLimit := middleware.NewRateLimiter(20, cfg)

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	r.Get("/signup", auth.SignUp)
	r.Post("/signup", loginLimit.Middleware(http.HandlerFunc(auth.SignUpPost)).ServeHTTP)
	r.Get("/forgot-password", auth.ForgotPassword)
	r.Post("/forgot-password", resetLimit.Middleware(http.HandlerFunc(auth.ForgotPasswordPost)).ServeHTTP)
	r.Get("/reset-password", auth.ResetPassword)
	r.Post("/reset-password", resetLimit.Middleware(http.HandlerFunc(auth.ResetPasswordPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)
	r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
}
`

const tplRoutesMinimal = `package app

import (
	"github.com/puppe1990/cais/pkg/cais"
	"{{.ModulePath}}/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)
	r.Get("/", home.ServeHTTP)
}
`

const tplMainBlank = `package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/boot"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"{{.ModulePath}}/internal/app"
	appi18n "{{.ModulePath}}/internal/i18n"
	"{{.ModulePath}}/internal/store"
	"{{.ModulePath}}/web"
)

func main() {
	cfg := cais.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	preferredPort := cfg.Port
	port, shifted, err := cais.ResolvePort(cfg.Port, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Port = port

	a, err := bootstrapWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	shiftedFrom := ""
	if shifted {
		shiftedFrom = preferredPort
	}
	boot.Print(os.Stdout, boot.Options{
		AppName:         "{{.AppName}}",
		Config:          cfg,
		Version:         boot.CaisVersion(),
		PortShiftedFrom: shiftedFrom,
	})
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func bootstrap() (*app.App, error) {
	return bootstrapWithConfig(cais.Load())
}

func bootstrapWithConfig(cfg cais.Config) (*app.App, error) {
	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	catalog := appi18n.NewCatalog(cfg.Locale)
	renderer, err := cais.NewRenderer(tmplFS, catalog)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := findWebDir("static")
	if err != nil {
		_ = s.Close()
		return nil, err
	}

	return app.New(cfg, app.Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: staticDir,
		Site:      meta.SiteFrom("{{.AppName}}", cfg.AppURL),
		Catalog:   catalog,
	})
}

func findWebDir(subpath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(wd, "web", subpath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("web/%s not found", subpath)
		}
		wd = parent
	}
}
`

const tplAppBlank = `package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"{{.ModulePath}}/internal/store"
)

type Deps struct {
	Renderer  *cais.Renderer
	Store     store.Store
	StaticDir string
	Site      meta.Site
	Catalog   *i18n.Catalog
}

type App struct {
	config cais.Config
	store  store.Store
	router *cais.Router
	server *http.Server
}

func New(cfg cais.Config, deps Deps) (*App, error) {
	if deps.Renderer == nil {
		return nil, fmt.Errorf("renderer is required")
	}
	if deps.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	site := deps.Site
	if site.AppName == "" {
		site = meta.SiteFrom("{{.AppName}}", cfg.AppURL)
	}
	deps.Site = site

	r := cais.NewRouter()
	r.Use(middleware.CSRF(cfg))
	buf := devlog.Prepare(cfg.Env)
	if buf != nil {
		r.Use(middleware.LoggerTo(cfg, devlog.MirrorDefault(log.Writer())))
	} else {
		r.Use(middleware.Logger(cfg))
	}
	r.Use(middleware.Recover)
	r.Use(middleware.SecurityHeaders(cfg))
	r.Static("/static", deps.StaticDir)

	registerRoutes(r, deps, cfg)
	devlog.Register(r, cfg.Env, buf)
	r.Get("/health", healthHandler(deps.Store))

	return &App{
		config: cfg,
		store:  deps.Store,
		router: r,
		server: &http.Server{
			Addr:              cfg.Port,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

func healthHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if err := s.Ping(); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}

func (a *App) Handler() http.Handler {
	return a.router
}

func (a *App) Run() error {
	return a.RunContext(context.Background())
}

func (a *App) RunContext(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			_ = a.store.Close()
			return err
		}
		<-errCh
		_ = a.store.Close()
		return nil
	case err := <-errCh:
		_ = a.store.Close()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
`

const tplRoutesBlank = `package app

import (
	"github.com/puppe1990/cais/pkg/cais"
	"{{.ModulePath}}/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)
	r.Get("/", home.ServeHTTP)
}
`
