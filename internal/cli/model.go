package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type modelOpts struct {
	Fields string
	dryRun bool
}

func parseModelOpts(args []string) (modelOpts, error) {
	opts := modelOpts{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--fields":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--fields requires a value")
			}
			i++
			opts.Fields = args[i]
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}

func scaffoldModel(dir, name string, opts modelOpts) error {
	fields, err := parseFields(opts.Fields)
	if err != nil {
		return err
	}

	data := dataForResource(name)
	data.ModulePath = readModulePath(dir)
	data.Fields = fields
	data.Seed = false

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	if !opts.dryRun {
		if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
			return err
		}
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
		filepath.Join("internal/models", data.Snake+".go"):                                   buildResourceModel(data),
		filepath.Join("internal/store/migrations", data.MigrationNum+"_"+data.Plural+".sql"): buildResourceMigration(data),
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeScaffoldFile(full, []byte(content), 0o644, path, opts.dryRun); err != nil {
			return err
		}
	}

	if err := patchStoreForResource(dir, data, opts.dryRun); err != nil {
		return err
	}
	if opts.dryRun {
		return nil
	}
	return gofmtGoFiles(dir)
}
