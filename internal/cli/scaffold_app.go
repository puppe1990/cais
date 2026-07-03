package cli

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed app_templates/supermarket/handler/*.tmpl
var supermarketHandlerFS embed.FS

//go:embed app_templates/supermarket/pages/*
var supermarketPagesFS embed.FS

//go:embed app_templates/supermarket/layout/*
var supermarketLayoutFS embed.FS

var appTemplateInstallers = map[string]func(string, scaffoldData, appScaffoldOpts) error{
	"supermarket": scaffoldAppSupermarket,
}

func listAppTemplates() []string {
	names := make([]string, 0, len(appTemplateInstallers))
	for name := range appTemplateInstallers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func scaffoldApp(dir, name string, opts appScaffoldOpts) error {
	install, ok := appTemplateInstallers[name]
	if !ok {
		return fmt.Errorf("unknown app template %q (available: %s)", name, strings.Join(listAppTemplates(), ", "))
	}
	data := scaffoldData{
		AppName:    filepath.Base(dir),
		ModulePath: moduleFromDir(dir),
	}
	return install(dir, data, opts)
}

func scaffoldAppSupermarket(dir string, data scaffoldData, opts appScaffoldOpts) error {
	if _, err := os.Stat(filepath.Join(dir, "internal/handlers/supermarket.go")); err == nil && !opts.force {
		return fmt.Errorf("supermarket app already installed — use --force to overwrite or add --data only after manual merge")
	}

	handlerFiles := map[string]string{
		"supermarket_test.go.tmpl": "internal/handlers/supermarket_test.go",
	}
	if opts.data {
		handlerFiles["supermarket_data.go.tmpl"] = "internal/handlers/supermarket.go"
	} else {
		handlerFiles["supermarket.go.tmpl"] = "internal/handlers/supermarket.go"
	}
	if err := writeEmbeddedDir(supermarketHandlerFS, "app_templates/supermarket/handler", dir, handlerFiles, opts.dryRun); err != nil {
		return err
	}

	if err := writeEmbeddedDir(supermarketPagesFS, "app_templates/supermarket/pages", dir, map[string]string{
		"scan.html":         "web/templates/pages/scan.html",
		"map.html":          "web/templates/pages/map.html",
		"feed.html":         "web/templates/pages/feed.html",
		"achievements.html": "web/templates/pages/achievements.html",
		"nfce.html":         "web/templates/pages/nfce.html",
	}, opts.dryRun); err != nil {
		return err
	}

	if err := writeEmbeddedFile(supermarketLayoutFS, "app_templates/supermarket/layout/base.html", filepath.Join(dir, "web/templates/layouts/base.html"), "web/templates/layouts/base.html", opts.dryRun); err != nil {
		return err
	}

	if err := patchRoutesForAppSupermarket(dir, opts.dryRun); err != nil {
		return err
	}
	if opts.data {
		if err := scaffoldAppSupermarketData(dir, data, opts.dryRun); err != nil {
			return err
		}
		return patchRoutesForAppSupermarketData(dir, opts.dryRun)
	}
	return nil
}

func writeEmbeddedFile(fsys embed.FS, src, dst, rel string, dryRun bool) error {
	body, err := fsys.ReadFile(src)
	if err != nil {
		return err
	}
	return writeScaffoldFile(dst, body, 0o644, rel, dryRun)
}

func writeEmbeddedDir(fsys embed.FS, root, dir string, mapping map[string]string, dryRun bool) error {
	for srcName, relPath := range mapping {
		src := root + "/" + srcName
		body, err := fsys.ReadFile(src)
		if err != nil {
			return fmt.Errorf("%s: %w", src, err)
		}
		dst := filepath.Join(dir, relPath)
		if err := writeScaffoldFile(dst, body, 0o644, relPath, dryRun); err != nil {
			return err
		}
	}
	return nil
}

func patchRoutesForAppSupermarket(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "NewSupermarketHandler") {
		return nil
	}

	content = strings.Replace(content,
		`home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)
`,
		`super := handlers.NewSupermarketHandler(deps.Renderer, deps.Site, cfg)
`,
		1)
	content = strings.Replace(content,
		`r.Get("/", home.ServeHTTP)`,
		`r.Get("/", super.Scan)
	r.Get("/map", super.Map)
	r.Get("/feed", super.Feed)
	r.Get("/achievements", super.Achievements)
	r.Get("/nfce", super.NFCe)`,
		1)

	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}
