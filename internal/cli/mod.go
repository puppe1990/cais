package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const frameworkModule = "github.com/puppe1990/cais"

func readModulePath(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func findLocalCaisReplace(appDir string) string {
	if p := os.Getenv("CAIS_REPLACE"); p != "" {
		return p
	}
	dir := appDir
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		for _, name := range []string{"Cais", "cais"} {
			candidate := filepath.Join(parent, name)
			if _, err := os.Stat(filepath.Join(candidate, "go.mod")); err != nil {
				continue
			}
			rel, err := filepath.Rel(appDir, candidate)
			if err != nil {
				continue
			}
			return rel
		}
		dir = parent
	}
	return ""
}

func patchGoModReplace(appDir string) error {
	replace := findLocalCaisReplace(appDir)
	if replace == "" {
		return nil
	}
	path := filepath.Join(appDir, "go.mod")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "replace "+frameworkModule) {
		return nil
	}
	block := fmt.Sprintf("\nreplace %s => %s\n", frameworkModule, replace)
	content = strings.TrimRight(content, "\n") + block
	return os.WriteFile(path, []byte(content), 0o644)
}
