package cais

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveWebDir returns the directory for web assets (static or templates).
// When override is set (STATIC_DIR / TEMPLATES_DIR), it is used as-is.
// Otherwise the path is discovered by walking up from the working directory.
func ResolveWebDir(subpath, override string) (string, error) {
	if override != "" {
		if _, err := os.Stat(override); err != nil {
			env := "STATIC_DIR"
			if subpath == "templates" {
				env = "TEMPLATES_DIR"
			}
			return "", fmt.Errorf("%s: %w (set %s or systemd WorkingDirectory to app root with web/%s)", override, err, env, subpath)
		}
		return override, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(wd, "web", subpath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("web/%s not found (set STATIC_DIR/TEMPLATES_DIR or WorkingDirectory for deploy)", subpath)
		}
		wd = parent
	}
}
