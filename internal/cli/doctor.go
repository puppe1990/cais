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
	Info     bool
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
		checkDeployLayout(dir),
		checkQualityTooling(dir),
	}
	if isProduction(dir) {
		checks = append(checks, checkAdminToken(dir), checkAppURL(dir))
		if hasAuthHandler(dir) {
			checks = append(checks, checkSMTP(dir))
		}
	}
	if c := checkSeedsInfo(dir); c != nil {
		checks = append(checks, *c)
	}

	var failed int
	for _, c := range checks {
		mark := "ok"
		if c.Info {
			mark = "info"
		} else if !c.OK {
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
		if !c.OK && !c.Info && c.FixHint != "" {
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

func checkDeployLayout(dir string) doctorCheck {
	static := filepath.Join(dir, "web", "static")
	manifest := filepath.Join(static, "manifest.webmanifest")
	if _, err := os.Stat(static); err != nil {
		return doctorCheck{
			Name:    "deploy layout",
			Detail:  "missing web/static",
			FixHint: "run cais css && make pwa; deploy needs web/static beside the binary",
		}
	}
	if _, err := os.Stat(manifest); err != nil {
		return doctorCheck{
			Name:    "deploy layout",
			Detail:  "missing manifest.webmanifest",
			FixHint: "run make pwa from the Cais framework or cais new",
		}
	}
	return doctorCheck{Name: "deploy layout", OK: true, Detail: "web/static ready for systemd deploy"}
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

func isProduction(dir string) bool {
	return resolveEnvVar(dir, "ENV") == "production"
}

func resolveEnvVar(dir, key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		return ""
	}
	return parseDotEnv(data)[key]
}

func parseDotEnv(data []byte) map[string]string {
	vars := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		vars[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return vars
}

func checkAdminToken(dir string) doctorCheck {
	if resolveEnvVar(dir, "ADMIN_TOKEN") != "" {
		return doctorCheck{Name: "ADMIN_TOKEN", OK: true}
	}
	return doctorCheck{
		Name:     "ADMIN_TOKEN",
		Optional: true,
		Detail:   "required when ENV=production",
		FixHint:  "set ADMIN_TOKEN in .env",
	}
}

func checkAppURL(dir string) doctorCheck {
	if resolveEnvVar(dir, "APP_URL") != "" {
		return doctorCheck{Name: "APP_URL", OK: true}
	}
	return doctorCheck{
		Name:     "APP_URL",
		Optional: true,
		Detail:   "required when ENV=production",
		FixHint:  "set APP_URL in .env",
	}
}

func hasAuthHandler(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "internal/handlers/auth.go"))
	return err == nil
}

func checkSMTP(dir string) doctorCheck {
	if resolveEnvVar(dir, "SMTP_HOST") != "" && resolveEnvVar(dir, "SMTP_FROM") != "" {
		return doctorCheck{Name: "SMTP", OK: true}
	}
	return doctorCheck{
		Name:     "SMTP",
		Optional: true,
		Detail:   "password reset emails log to stdout without SMTP_HOST/SMTP_FROM",
		FixHint:  "set SMTP_HOST and SMTP_FROM in .env for outbound mail",
	}
}

func checkSeedsInfo(dir string) *doctorCheck {
	path := filepath.Join(dir, "internal/db/seeds.go")
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	return &doctorCheck{
		Name:   "db seeds",
		OK:     true,
		Info:   true,
		Detail: "run cais db seed for catalog data (idempotent; safe in production)",
	}
}
