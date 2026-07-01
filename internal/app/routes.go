package app

import (
	"github.com/puppe1990/cais/internal/handlers"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config, site meta.Site) {
	home := handlers.NewHomeHandler(deps.Renderer, site)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, site)
	dashboard := handlers.NewDashboardHandler(deps.Renderer, deps.Store, site)
	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, site, deps.Store.Sessions())

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contact.Post)
	r.Get("/login", auth.Login)
	r.Post("/login", auth.LoginPost)
	r.Post("/logout", auth.LogoutPost)
	r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
}
