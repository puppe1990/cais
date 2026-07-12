package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

// hasViteApp reports whether dir is an Inertia + Vite app (vite.config.js + npm build script).
func hasViteApp(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "vite.config.js")); err != nil {
		return false
	}
	body, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(body, &pkg); err != nil {
		return false
	}
	_, ok := pkg.Scripts["build"]
	return ok
}

func runViteBuild(dir string) error {
	if !hasViteApp(dir) {
		return nil
	}
	return runCmd(dir, "npm", "run", "build")
}

func startViteWatch(dir string) (*exec.Cmd, error) {
	if !hasViteApp(dir) {
		return nil, nil
	}
	cmd := exec.Command("npx", "vite", "build", "--watch")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
