// internal/app/app.go templates (full and blank cais new).
package cli

const tplApp = `package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	inertia "github.com/romsar/gonertia/v3"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/cais/pkg/cais/netutil"
	"{{.ModulePath}}/internal/store"
)

type Deps struct {
	Renderer  *cais.Renderer
	Store     store.Store
	StaticDir string
	Site      meta.Site
	Catalog   *i18n.Catalog
	Inertia   *inertia.Inertia
}

type App struct {
	config cais.Config
	store  store.Store
	router *cais.Router
	server *http.Server
}

const defaultInertiaRoot = ` + "`" + `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8" />{{"{{ .inertiaHead }}"}}</head>
<body>{{"{{ .inertia }}"}}</body>
</html>` + "`" + `

func New(cfg cais.Config, deps Deps) (*App, error) {
	if deps.Renderer == nil {
		return nil, fmt.Errorf("renderer is required")
	}
	if deps.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	inertiaI := deps.Inertia
	if inertiaI == nil {
		var err error
		inertiaI, err = inertia.New(defaultInertiaRoot)
		if err != nil {
			return nil, fmt.Errorf("inertia: %w", err)
		}
	}

	site := deps.Site
	if site.AppName == "" {
		site = meta.SiteFrom("{{.AppName}}", cfg.AppURL)
	}
	site.Env = cfg.Env
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
	r.StaticForEnv("/static", deps.StaticDir, cfg)

	registerRoutes(r, deps, cfg)
	devlog.Register(r, cfg.Env, buf)
	r.Get("/health", healthHandler(deps.Store, cfg))

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

func healthHandler(s store.Store, cfg cais.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if err := s.Ping(); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(netutil.HealthPayload(status, cfg.Port))
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

const tplAppBlank = tplApp
