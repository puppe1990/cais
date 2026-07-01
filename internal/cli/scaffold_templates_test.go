package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldHandler_usesModernHandlerPattern(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "pageapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "pageapp",
		ModulePath: "github.com/puppe1990/pageapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldHandler(appDir, "about", false); err != nil {
		t.Fatal(err)
	}

	handler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/about.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(handler)
	for _, needle := range []string{
		"httpx.RenderOrError",
		"meta.ForRequest",
		"meta.Site",
		"i18n.Catalog",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("about.go missing %q", needle)
		}
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routes), "handlers.NewAboutHandler(deps.Renderer, deps.Site, deps.Catalog, cfg)") {
		t.Error("routes.go should wire site, catalog, and cfg into handler")
	}
}

func TestScaffoldResource_adminFormUsesFormHelpers(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "adminapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "adminapp",
		ModulePath: "github.com/puppe1990/adminapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "widget", resourceOpts{
		Fields: "name:string,url:url",
	}); err != nil {
		t.Fatal(err)
	}

	form, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/admin_widget_form.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(form)
	if !strings.Contains(body, `{{ csrfField .CSRFToken }}`) {
		t.Error("admin form should use csrfField helper")
	}
	if !strings.Contains(body, `{{ fieldInput (makeField`) {
		t.Error("admin form should use fieldInput/makeField helpers")
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_widgets.go"))
	if err != nil {
		t.Fatal(err)
	}
	adminBody := string(admin)
	if !strings.Contains(adminBody, "validate.FieldErrors") {
		t.Error("admin handler should use validate.FieldErrors")
	}
	if !strings.Contains(adminBody, "StatusUnprocessableEntity") {
		t.Error("admin handler should return 422 on validation errors")
	}
}
