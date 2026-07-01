package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldResource(dir, name string) error {
	data := dataForResource(name)
	data.ModulePath = readModulePath(dir)

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}
	sqlCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlCount++
		}
	}
	data.MigrationNum = fmt.Sprintf("%03d", sqlCount+1)

	files := map[string]string{
		filepath.Join("internal/models", data.Snake+".go"):                                   tplResourceModel,
		filepath.Join("internal/handlers", "admin_"+data.Plural+".go"):                       tplResourceAdminHandler,
		filepath.Join("internal/handlers", "admin_"+data.Plural+"_test.go"):                  tplResourceAdminTest,
		filepath.Join("web/templates/pages", "admin_"+data.Plural+".html"):                   tplResourceAdminIndex,
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_form.html"):               tplResourceAdminForm,
		filepath.Join("internal/store/migrations", data.MigrationNum+"_"+data.Plural+".sql"): tplResourceMigration,
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeTemplate(full, content, data); err != nil {
			return err
		}
		_, _ = fmt.Printf("  create %s\n", path)
	}

	if err := patchStoreForResource(dir, data); err != nil {
		return err
	}
	if err := patchStoreTestForResource(dir, data); err != nil {
		return err
	}
	return patchRoutesForResource(dir, data)
}

func patchStoreForResource(dir string, data scaffoldData) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "Insert"+data.Pascal) {
		return nil
	}

	ifaceMarker := "\n\tClose() error"
	if !strings.Contains(content, ifaceMarker) {
		return fmt.Errorf("could not patch store interface")
	}
	ifaceInsert := fmt.Sprintf(
		"\n\tInsert%s(models.%s) (int64, error)\n\tUpdate%s(models.%s) error\n\tDelete%s(id int64) error\n\tFind%sByID(id int64) (models.%s, error)\n\tListAll%s() ([]models.%s, error)",
		data.Pascal, data.Pascal,
		data.Pascal, data.Pascal,
		data.Pascal,
		data.Pascal, data.Pascal,
		data.PluralPascal, data.Pascal,
	)
	content = strings.Replace(content, ifaceMarker, ifaceInsert+ifaceMarker, 1)

	implMarker := "\nfunc (s *SQLiteStore) Close()"
	if !strings.Contains(content, implMarker) {
		return fmt.Errorf("could not patch store implementation")
	}
	implInsert, err := renderSnippet(tplResourceStoreMethods, data)
	if err != nil {
		return err
	}
	content = strings.Replace(content, implMarker, implInsert+implMarker, 1)

	if !strings.Contains(content, data.ModulePath+"/internal/models") {
		content = strings.Replace(content,
			`"database/sql"`,
			`"database/sql"

	"`+data.ModulePath+`/internal/models"`,
			1,
		)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update internal/store/store.go")
	return nil
}

func patchStoreTestForResource(dir string, data scaffoldData) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "TestStore_Insert"+data.Pascal) {
		return nil
	}

	insert, err := renderSnippet(tplResourceStoreTest, data)
	if err != nil {
		return err
	}
	if !strings.Contains(content, data.ModulePath+"/internal/models") {
		content = strings.Replace(content,
			`import "testing"`,
			`import (
	"testing"

	"`+data.ModulePath+`/internal/models"
)`,
			1,
		)
	}
	content = strings.TrimRight(content, "\n") + "\n" + insert
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update internal/store/store_test.go")
	return nil
}

func patchRoutesForResource(dir string, data scaffoldData) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "/admin/"+data.Plural) {
		return nil
	}

	if !strings.Contains(content, "github.com/matheuspuppe/cais/pkg/cais/middleware") {
		content = strings.Replace(content,
			`"github.com/matheuspuppe/cais/pkg/cais"`,
			`"github.com/matheuspuppe/cais/pkg/cais"
	"github.com/matheuspuppe/cais/pkg/cais/middleware"`,
			1,
		)
	}

	insert := fmt.Sprintf(`
	admin%s := handlers.NewAdmin%sHandler(deps.Renderer, deps.Store)
	r.Get("/admin/%s", middleware.Protect(admin%s.Index))
	r.Get("/admin/%s/new", middleware.Protect(admin%s.New))
	r.Post("/admin/%s", middleware.Protect(admin%s.Create))
	r.Get("/admin/%s/{id}/edit", middleware.Protect(cais.IntParam("id", admin%s.Edit)))
	r.Post("/admin/%s/{id}", middleware.Protect(cais.IntParam("id", admin%s.Update)))
	r.Post("/admin/%s/{id}/delete", middleware.Protect(cais.IntParam("id", admin%s.Delete)))
`,
		data.PluralPascal, data.PluralPascal,
		data.Plural, data.PluralCamel,
		data.Plural, data.PluralCamel,
		data.Plural, data.PluralCamel,
		data.Plural, data.PluralCamel,
		data.Plural, data.PluralCamel,
		data.Plural, data.PluralCamel,
	)

	marker := "\n}\n"
	idx := strings.LastIndex(content, marker)
	if idx == -1 {
		return fmt.Errorf("could not patch routes.go")
	}
	content = content[:idx] + insert + content[idx:]
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update internal/app/routes.go")
	return nil
}
