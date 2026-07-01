package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func scaffoldConsole(appDir string, dryRun bool) error {
	rel := "cmd/console/main.go"
	path := filepath.Join(appDir, rel)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("cmd/console/main.go already exists")
	}

	data := appScaffoldData(appDir)
	if data.ModulePath == "" {
		return fmt.Errorf("could not read module path from go.mod")
	}

	if err := writeScaffoldTemplate(path, tplConsole, data, rel, dryRun); err != nil {
		return err
	}
	if !dryRun {
		_, _ = fmt.Println("  create cmd/console/main.go")
	}
	return nil
}

func appScaffoldData(appDir string) scaffoldData {
	mod := readModulePath(appDir)
	name := filepath.Base(appDir)
	return scaffoldData{AppName: name, ModulePath: mod}
}
