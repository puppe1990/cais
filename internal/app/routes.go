package app

import (
	"net/http"

	"github.com/puppe1990/cais/internal/handlers"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config, site meta.Site) {
	home := handlers.NewHomeHandler(deps.Renderer, site, deps.Catalog, cfg)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, site, deps.Catalog, cfg)
	dashboard := handlers.NewDashboardHandler(deps.Renderer, deps.Store, site, cfg)
	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, site, deps.Store.Sessions(), cfg, deps.Catalog)

	loginLimit := middleware.NewRateLimiter(10, cfg)
	contactLimit := middleware.NewRateLimiter(20, cfg)

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)
	r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
}
