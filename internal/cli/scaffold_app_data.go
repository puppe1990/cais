package cli

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed app_templates/supermarket/data/*
var supermarketDataFS embed.FS

const tplStoreMercadoInterfaceMethods = `	LoadStats(userID int64) (level, points, rank int, err error)
	ListProducts(limit int) ([]models.Product, error)
	FindProductByBarcode(barcode string) (models.Product, bool, error)
	CreateProduct(name, barcode, category string) (int64, error)
	ProductAvgPriceCents(productID int64) (int, error)
	ListSupermarkets() ([]models.Supermarket, error)
	ListFeedReports(limit int) ([]models.PriceReport, error)
	ListUserReports(userID int64, limit int) ([]models.PriceReport, error)
	CreatePriceReport(userID, productID, supermarketID int64, priceCents int) (int64, error)
	ConfirmPriceReport(reportID, userID int64) (int, error)
	FlagPriceReport(reportID int64) error
	ListBadges(userID int64) ([]models.Badge, error)
	Leaderboard(limit int, currentUserID int64) ([]models.LeaderboardEntry, error)
	SupermarketOfferCount(supermarketID int64) (int, error)
	SupermarketBestDeal(supermarketID int64) (string, error)
	SeedMercadoDemo() error
`

const tplStoreMercadoBootSeed = `	st := &SQLiteStore{db: wrapped}
	if env == "development" {
		if err := st.SeedMercadoDemo(); err != nil {
			_ = wrapped.Close()
			return nil, err
		}
	}
	return st, nil`

func scaffoldAppSupermarketData(dir string, data scaffoldData, dryRun bool) error {
	migrationPath, _, err := nextMigrationFile(dir, "mercado", dryRun)
	if err != nil {
		return err
	}
	migrationBody, err := supermarketDataFS.ReadFile("app_templates/supermarket/data/003_mercado.sql")
	if err != nil {
		return err
	}
	if err := writeScaffoldFile(filepath.Join(dir, migrationPath), migrationBody, 0o644, migrationPath, dryRun); err != nil {
		return err
	}

	pairs := []struct{ src, dst string }{
		{"app_templates/supermarket/data/product.go.tmpl", "internal/models/product.go"},
		{"app_templates/supermarket/data/supermarket.go.tmpl", "internal/models/supermarket.go"},
		{"app_templates/supermarket/data/mercado_models.go.tmpl", "internal/models/mercado.go"},
		{"app_templates/supermarket/data/mercado_store.go.tmpl", "internal/store/mercado.go"},
		{"app_templates/supermarket/data/mercado_test.go.tmpl", "internal/store/mercado_test.go"},
		{"app_templates/supermarket/data/mercado_seed.go.tmpl", "internal/store/mercado_seed.go"},
		{"app_templates/supermarket/data/feed_confirm_btn.html", "web/templates/partials/feed_confirm_btn.html"},
	}
	for _, p := range pairs {
		body, err := supermarketDataFS.ReadFile(p.src)
		if err != nil {
			return fmt.Errorf("%s: %w", p.src, err)
		}
		body = rewriteModulePath(body, data.ModulePath)
		if err := writeScaffoldFile(filepath.Join(dir, p.dst), body, 0o644, p.dst, dryRun); err != nil {
			return fmt.Errorf("%s: %w", p.dst, err)
		}
	}
	if err := patchStoreForMercado(dir, dryRun); err != nil {
		return err
	}
	return patchAppForMercadoStats(dir, dryRun)
}

func rewriteModulePath(body []byte, module string) []byte {
	return []byte(strings.ReplaceAll(string(body), "github.com/puppe1990/mercado", module))
}

func patchStoreForMercado(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "SeedMercadoDemo") {
		return nil
	}
	marker := "\tSessions() session.Store\n"
	if !strings.Contains(content, marker) {
		return fmt.Errorf("could not patch store.go: missing Sessions() marker")
	}
	insert := marker + tplStoreMercadoInterfaceMethods
	content = strings.Replace(content, marker, insert, 1)
	content = strings.Replace(content, `return &SQLiteStore{db: wrapped}, nil`, tplStoreMercadoBootSeed, 1)
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func patchAppForMercadoStats(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/app.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "LoadUserStats") {
		return nil
	}
	content = strings.Replace(content,
		`r.Use(middleware.LoadSession(deps.Store.Sessions()))`,
		`r.Use(middleware.LoadSession(deps.Store.Sessions()))
	r.Use(middleware.LoadUserStats(deps.Store))`,
		1,
	)
	return updateScaffoldFile(path, []byte(content), "internal/app/app.go", dryRun)
}

func patchRoutesForAppSupermarketData(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "super.LookupPost") {
		return nil
	}
	content = strings.Replace(content,
		`super := handlers.NewSupermarketHandler(deps.Renderer, deps.Site, cfg)`,
		`super := handlers.NewSupermarketHandler(deps.Renderer, deps.Store, deps.Site, cfg)`,
		1,
	)
	if !strings.Contains(content, "actionLimit") {
		content = strings.Replace(content,
			`contactLimit := middleware.NewRateLimiter(20, cfg)`,
			`contactLimit := middleware.NewRateLimiter(20, cfg)
	actionLimit := middleware.NewRateLimiter(30, cfg)`,
			1,
		)
	}
	insert := `	r.Post("/scan/lookup", actionLimit.Middleware(http.HandlerFunc(super.LookupPost)).ServeHTTP)
	r.Post("/scan/report", actionLimit.Middleware(http.HandlerFunc(super.ReportPost)).ServeHTTP)
`
	if !strings.Contains(content, "super.LookupPost") {
		content = strings.Replace(content, `r.Get("/", super.Scan)
`, `r.Get("/", super.Scan)
`+insert, 1)
	}
	if !strings.Contains(content, "super.ConfirmPost") {
		content = strings.Replace(content, `r.Get("/feed", super.Feed)
`, `r.Get("/feed", super.Feed)
	r.Post("/feed/{id}/confirm", cais.IntParam("id", super.ConfirmPost))
	r.Post("/feed/{id}/flag", cais.IntParam("id", super.FlagPost))
`, 1)
	}
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}
