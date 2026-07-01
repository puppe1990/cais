package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const makefileCIMarker = "# cais quality tooling"

const tplMakefileCIBlock = makefileCIMarker + `
.PHONY: lint format format-check pre-commit-install ci

lint:
	golangci-lint run ./...

format:
	npm run format

format-check:
	npm run format:check

pre-commit-install:
	pre-commit install

ci: test lint format-check
`

func scaffoldCI(dir string, data scaffoldData, dryRun bool) error {
	if data.ModulePath == "" {
		data.ModulePath = moduleFromDir(dir)
	}

	created := 0
	for path, content := range qualityToolingFiles() {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			continue
		}
		if err := writeScaffoldTemplate(full, content, data, path, dryRun); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if !dryRun {
			_, _ = fmt.Printf("  create %s\n", path)
		}
		created++
	}

	if err := patchMakefileForCI(dir, dryRun); err != nil {
		return err
	}
	if err := patchPackageJSONForCI(dir, dryRun); err != nil {
		return err
	}

	if !dryRun {
		if created == 0 {
			_, _ = fmt.Println("  quality tooling already present (Makefile and package.json patched if needed)")
		}
		_, _ = fmt.Println("\nNext: make pre-commit-install && make ci")
	}
	return nil
}

func patchMakefileForCI(dir string, dryRun bool) error {
	path := filepath.Join(dir, "Makefile")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if dryRun {
				printfScaffold("create", "Makefile")
				return nil
			}
			if err := os.WriteFile(path, []byte(tplMakefile), 0o644); err != nil {
				return err
			}
			_, _ = fmt.Println("  create Makefile")
			return nil
		}
		return err
	}
	if makefileHasCITargets(string(body)) {
		return nil
	}
	updated := strings.TrimRight(string(body), "\n") + "\n" + tplMakefileCIBlock
	if dryRun {
		printfScaffold("update", "Makefile")
		return nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update Makefile")
	return nil
}

func patchPackageJSONForCI(dir string, dryRun bool) error {
	path := filepath.Join(dir, "package.json")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, `"test"`) {
		return nil
	}
	const needle = `"format:check": "prettier --check ."`
	if !strings.Contains(content, needle) {
		return fmt.Errorf("package.json: add scripts.test manually (expected %s)", needle)
	}
	updated := strings.Replace(content, needle,
		needle+",\n    \"test\": \"npm run format:check\"", 1)
	if dryRun {
		printfScaffold("update", "package.json")
		return nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Println("  update package.json")
	return nil
}

func makefileHasCITargets(body string) bool {
	return strings.Contains(body, makefileCIMarker) || strings.Contains(body, "\nci:")
}
