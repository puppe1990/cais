// cmd/server/main.go templates (full and blank cais new).
package cli

const tplMain = `package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

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
	templatesDir, err := cais.ResolveWebDir("templates", cfg.TemplatesDir)
	if err != nil {
		templatesDir = ""
	}
	renderer, err := cais.NewRendererForEnv(cfg, tmplFS, templatesDir, catalog)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := cais.ResolveWebDir("static", cfg.StaticDir)
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
`

const tplMainBlank = `package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

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
	templatesDir, err := cais.ResolveWebDir("templates", cfg.TemplatesDir)
	if err != nil {
		templatesDir = ""
	}
	renderer, err := cais.NewRendererForEnv(cfg, tmplFS, templatesDir, catalog)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := cais.ResolveWebDir("static", cfg.StaticDir)
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
`
