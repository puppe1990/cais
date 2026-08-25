package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// viteMainJSRel is the Inertia entry app.html always loads. Checking the build
// directory is not enough: scaffold writes web/static/build/.gitkeep (#159).
const viteMainJSRel = "web/static/build/assets/main.js"

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
	if err := runCmd(dir, "npm", "run", "build"); err != nil {
		return err
	}
	return requireViteMainJS(dir)
}

func requireViteMainJS(dir string) error {
	path := filepath.Join(dir, viteMainJSRel)
	if _, err := os.Stat(path); err != nil {
		hint := "npm run build did not emit " + viteMainJSRel
		if _, nmErr := os.Stat(filepath.Join(dir, "node_modules")); nmErr != nil {
			return fmt.Errorf("%s: %w (run cais install if node_modules is missing)", hint, err)
		}
		return fmt.Errorf("%s: %w", hint, err)
	}
	return nil
}

// viteWatchArgs is the npm invocation used by cais dev to rebuild SPA assets
// when web/src/** changes. Uses the project's package.json "build" script so the
// local vite version is always used (not a global/npx mismatch).
func viteWatchArgs() []string {
	return []string{"run", "build", "--", "--watch"}
}

func startViteWatch(dir string) (*exec.Cmd, error) {
	if !hasViteApp(dir) {
		return nil, nil
	}
	if _, err := os.Stat(filepath.Join(dir, "node_modules")); err != nil {
		return nil, fmt.Errorf("node_modules missing — run cais install before cais dev")
	}
	cmd := exec.Command("npm", viteWatchArgs()...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
