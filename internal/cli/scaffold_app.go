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

var appTemplateInstallers = map[string]func(string, scaffoldData, bool) error{
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

func scaffoldApp(dir, name string, dryRun bool) error {
	install, ok := appTemplateInstallers[name]
	if !ok {
		return fmt.Errorf("unknown app template %q (available: %s)", name, strings.Join(listAppTemplates(), ", "))
	}
	data := scaffoldData{
		AppName:    filepath.Base(dir),
		ModulePath: moduleFromDir(dir),
	}
	return install(dir, data, dryRun)
}

func scaffoldAppSupermarket(dir string, data scaffoldData, dryRun bool) error {
	if _, err := os.Stat(filepath.Join(dir, "internal/handlers/supermarket.go")); err == nil {
		return fmt.Errorf("supermarket app already installed — remove internal/handlers/supermarket.go first")
	}

	if err := writeEmbeddedDir(supermarketHandlerFS, "app_templates/supermarket/handler", dir, map[string]string{
		"supermarket.go.tmpl":      "internal/handlers/supermarket.go",
		"supermarket_test.go.tmpl": "internal/handlers/supermarket_test.go",
	}, dryRun); err != nil {
		return err
	}

	if err := writeEmbeddedDir(supermarketPagesFS, "app_templates/supermarket/pages", dir, map[string]string{
		"scan.html":         "web/templates/pages/scan.html",
		"map.html":          "web/templates/pages/map.html",
		"feed.html":         "web/templates/pages/feed.html",
		"achievements.html": "web/templates/pages/achievements.html",
		"nfce.html":         "web/templates/pages/nfce.html",
	}, dryRun); err != nil {
		return err
	}

	if err := writeEmbeddedFile(supermarketLayoutFS, "app_templates/supermarket/layout/base.html", filepath.Join(dir, "web/templates/layouts/base.html"), "web/templates/layouts/base.html", dryRun); err != nil {
		return err
	}

	return patchRoutesForAppSupermarket(dir, dryRun)
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
