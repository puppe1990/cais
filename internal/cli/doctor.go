package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type doctorCheck struct {
	Name     string
	OK       bool
	Optional bool
	Detail   string
	FixHint  string
}

func runDoctor(w io.Writer, dir string) error {
	checks := []doctorCheck{
		checkGoMod(dir),
		checkCaisDep(dir),
		checkHTMX(dir),
		checkAir(),
		checkCSS(dir),
		checkQualityTooling(dir),
	}

	var failed int
	for _, c := range checks {
		mark := "ok"
		if !c.OK {
			if c.Optional {
				mark = "warn"
			} else {
				mark = "FAIL"
				failed++
			}
		}
		_, _ = fmt.Fprintf(w, "[%s] %s", mark, c.Name)
		if c.Detail != "" {
			_, _ = fmt.Fprintf(w, " — %s", c.Detail)
		}
		_, _ = fmt.Fprintln(w)
		if !c.OK && c.FixHint != "" {
			_, _ = fmt.Fprintf(w, "      fix: %s\n", c.FixHint)
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}
	_, _ = fmt.Fprintln(w, "All checks passed.")
	return nil
}

func checkGoMod(dir string) doctorCheck {
	path := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{Name: "go.mod", FixHint: "run from a Cais app root"}
	}
	return doctorCheck{Name: "go.mod", OK: true}
}

func checkCaisDep(dir string) doctorCheck {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return doctorCheck{Name: "cais dependency", Detail: err.Error()}
	}
	content := string(data)
	if !strings.Contains(content, frameworkModule) {
		return doctorCheck{Name: "cais dependency", Detail: "missing " + frameworkModule, FixHint: "cais new or add require in go.mod"}
	}
	if strings.Contains(content, "replace "+frameworkModule) {
		return doctorCheck{Name: "cais dependency", OK: true, Detail: "local replace active"}
	}
	return doctorCheck{Name: "cais dependency", OK: true, Detail: "v" + extractCaisVersion(content)}
}

func extractCaisVersion(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, frameworkModule) && strings.Contains(line, "v") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "v") {
					return strings.TrimPrefix(p, "v")
				}
			}
		}
	}
	return "?"
}

func checkHTMX(dir string) doctorCheck {
	path := filepath.Join(dir, "web/static/js/htmx.min.js")
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{
			Name:    "htmx.min.js",
			Detail:  "missing",
			FixHint: "re-run cais new or copy from Cais web/static/js/htmx.min.js",
		}
	}
	return doctorCheck{Name: "htmx.min.js", OK: true}
}

func checkAir() doctorCheck {
	if path, err := exec.LookPath("air"); err == nil {
		return doctorCheck{Name: "air", OK: true, Detail: path}
	}
	home, _ := os.UserHomeDir()
	candidate := filepath.Join(home, "go/bin/air")
	if _, err := os.Stat(candidate); err == nil {
		return doctorCheck{Name: "air", OK: true, Detail: candidate}
	}
	return doctorCheck{
		Name:     "air",
		Optional: true,
		Detail:   "not found",
		FixHint:  "go install github.com/air-verse/air@latest",
	}
}

func checkQualityTooling(dir string) doctorCheck {
	path := filepath.Join(dir, ".github/workflows/ci.yml")
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{
			Name:     "quality tooling",
			Optional: true,
			Detail:   "CI/pre-commit not configured",
			FixHint:  "cais g ci",
		}
	}
	return doctorCheck{Name: "quality tooling", OK: true}
}

func checkCSS(dir string) doctorCheck {
	path := filepath.Join(dir, "web/static/css/styles.css")
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{Name: "tailwind css", Detail: "styles.css missing", FixHint: "cais install && cais css"}
	}
	return doctorCheck{Name: "tailwind css", OK: true}
}
