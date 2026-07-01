package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func scaffoldConsole(appDir string) error {
	path := filepath.Join(appDir, "cmd", "console", "main.go")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("cmd/console/main.go already exists")
	}

	data := appScaffoldData(appDir)
	if data.ModulePath == "" {
		return fmt.Errorf("could not read module path from go.mod")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := writeTemplate(path, tplConsole, data); err != nil {
		return err
	}
	_, _ = fmt.Println("  create cmd/console/main.go")
	return nil
}

func appScaffoldData(appDir string) scaffoldData {
	mod := readModulePath(appDir)
	name := filepath.Base(appDir)
	return scaffoldData{AppName: name, ModulePath: mod}
}
