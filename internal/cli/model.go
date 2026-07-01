package cli

import (
	"fmt"
	"os"
	"path/filepath"
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

	migrationPath, migrationNum, err := nextMigrationFile(dir, data.Plural, opts.dryRun)
	if err != nil {
		return err
	}
	data.MigrationNum = migrationNum

	files := map[string]string{
		filepath.Join("internal/models", data.Snake+".go"): buildResourceModel(data),
		migrationPath: buildResourceMigration(data),
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

	if err := patchStoreForResource(dir, data, opts.dryRun, false); err != nil {
		return err
	}
	if opts.dryRun {
		return nil
	}
	return gofmtGoFiles(dir)
}
