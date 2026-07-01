package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/puppe1990/cais/internal/handlers"
	"github.com/puppe1990/cais/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
)

type Deps struct {
	Renderer  *cais.Renderer
	Store     store.Store
	StaticDir string
}

type App struct {
	config cais.Config
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

	r := cais.NewRouter()
	r.Use(middleware.CSRF)
	buf := devlog.Prepare(cfg.Env)
	if buf != nil {
		r.Use(middleware.LoggerTo(devlog.MirrorDefault(log.Writer())))
	} else {
		r.Use(middleware.Logger)
	}
	devlog.Register(r, cfg.Env, buf)
	r.Use(middleware.Recover)
	r.Static("/static", deps.StaticDir)

	site := meta.SiteFrom("Cais", cfg.AppURL)
	home := handlers.NewHomeHandler(deps.Renderer, site)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, site)

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contact.Post)
	r.Get("/health", healthHandler)

	return &App{
		config: cfg,
		router: r,
		server: &http.Server{
			Addr:    cfg.Port,
			Handler: r,
		},
	}, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
			return err
		}
		<-errCh
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
