package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

type doctorOptions struct {
	Mobile bool
}

func runDoctor(w io.Writer, dir string, opts doctorOptions) error {
	checks := []doctorCheck{
		checkGoMod(dir),
		checkCaisDep(dir),
		checkCLIVersion(dir),
	}
	if isInertiaApp(dir) {
		checks = append(checks, checkInertiaFrontend(dir), checkViteConfig(dir))
	} else {
		checks = append(checks, checkHTMX(dir), checkSSEExt(dir))
	}
	checks = append(checks,
		checkSSEWriteTimeout(dir),
		checkAir(),
		checkCSS(dir),
		checkDeployLayout(dir),
		checkQualityTooling(dir),
	)
	if isProduction(dir) {
		checks = append(checks, checkAdminToken(dir), checkAppURL(dir))
		if hasAuthHandler(dir) {
			checks = append(checks, checkSMTP(dir))
		}
	}
	if c := checkSeedsInfo(dir); c != nil {
		checks = append(checks, *c)
	}
	if c := checkLocalCaisReplace(dir); c != nil {
		checks = append(checks, *c)
	}
	if opts.Mobile {
		checks = append(checks,
			checkFlashTemplate(dir),
			checkGoogleFonts(dir),
			checkPWACacheVersion(dir),
			checkChatSSEPattern(dir),
			checkSSEReconnectJS(dir),
			checkChatAgentJS(dir),
			checkChatEnterSubmitJS(dir),
			checkChatFormCSS(dir),
			checkChatScrollContainer(dir),
			checkHealthLANURLs(dir),
		)
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

func checkLocalCaisReplace(dir string) *doctorCheck {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil || !strings.Contains(string(data), "replace "+frameworkModule) {
		return nil
	}
	// cais link is for local work. Committing the replace breaks CI clones (#154).
	optional := !runningInCI()
	return &doctorCheck{
		Name:     "cais replace",
		Optional: optional,
		Detail:   "go.mod has a local cais replace — CI clones will not see that path",
		FixHint:  "cais link --unlink   # before commit/push",
	}
}

func runningInCI() bool {
	return os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"
}

func extractCaisVersion(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		// Prefer require line, skip replace.
		if strings.HasPrefix(line, "replace ") {
			continue
		}
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

// checkCLIVersion warns when the installed cais binary is older than go.mod
// (or too old for Vite watch). Optional: does not fail doctor.
func checkCLIVersion(dir string) doctorCheck {
	const name = "cais CLI version"
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: "skipped (no go.mod)"}
	}
	content := string(data)
	cliRaw := frameworkVersion()
	cli := parseSemverCore(cliRaw)
	if !cli.OK {
		// Source/dev build — assume current.
		return doctorCheck{Name: name, OK: true, Detail: "CLI " + cliRaw + " (dev build)"}
	}

	if strings.Contains(content, "replace "+frameworkModule) {
		return doctorCheck{
			Name:   name,
			OK:     true,
			Detail: "CLI v" + formatSemver(cli) + " (go.mod has local replace)",
		}
	}

	modRaw := extractCaisVersion(content)
	mod := parseSemverCore(modRaw)
	if mod.OK && compareSemverCore(cli, mod) < 0 {
		return doctorCheck{
			Name:     name,
			Optional: true,
			Detail:   fmt.Sprintf("CLI v%s is older than go.mod v%s — generators and cais dev may lack features", formatSemver(cli), formatSemver(mod)),
			FixHint:  fmt.Sprintf("go install %s/cmd/cais@v%s", frameworkModule, formatSemver(mod)),
		}
	}

	// Vite apps need CLI ≥ 0.8.0 for cais dev vite build --watch (#128).
	if hasViteApp(dir) {
		floor := parseSemverCore(minViteWatchVersion)
		if compareSemverCore(cli, floor) < 0 {
			return doctorCheck{
				Name:     name,
				Optional: true,
				Detail:   fmt.Sprintf("CLI v%s predates vite build --watch (need ≥ v%s) — SPA will not rebuild on web/src changes", formatSemver(cli), minViteWatchVersion),
				FixHint:  fmt.Sprintf("go install %s/cmd/cais@v%s  # or @latest after release", frameworkModule, minViteWatchVersion),
			}
		}
	}

	detail := "CLI v" + formatSemver(cli)
	if mod.OK {
		detail += " · go.mod v" + formatSemver(mod)
	}
	return doctorCheck{Name: name, OK: true, Detail: detail}
}

func formatSemver(s semverCore) string {
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}

func isInertiaApp(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "vite.config.js"))
	return err == nil
}

func checkInertiaFrontend(dir string) doctorCheck {
	appHTML := filepath.Join(dir, "web/templates/app.html")
	data, err := os.ReadFile(appHTML)
	if err != nil {
		return doctorCheck{
			Name:    "Inertia frontend",
			Detail:  "missing web/templates/app.html",
			FixHint: "re-run cais new or restore app.html from Cais Inertia scaffold",
		}
	}
	content := string(data)
	missing := []string{}
	for _, want := range []string{`{{ .inertia }}`, `{{ .inertiaHead }}`, `/static/build`} {
		if !strings.Contains(content, want) {
			missing = append(missing, want)
		}
	}
	mainJS := filepath.Join(dir, "web/src/main.js")
	if _, err := os.Stat(mainJS); err != nil {
		missing = append(missing, "web/src/main.js")
	}
	pagesDir := filepath.Join(dir, "web/src/pages")
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		missing = append(missing, "web/src/pages/")
	} else {
		hasSvelte := false
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".svelte") {
				hasSvelte = true
				break
			}
		}
		if !hasSvelte {
			missing = append(missing, "web/src/pages/*.svelte")
		}
	}
	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil || !strings.Contains(string(gomod), "gonertia") {
		missing = append(missing, "gonertia in go.mod")
	}
	if len(missing) > 0 {
		return doctorCheck{
			Name:    "Inertia frontend",
			Detail:  "missing: " + strings.Join(missing, ", "),
			FixHint: "cais install && npm run build; ensure gonertia is in go.mod",
		}
	}
	return doctorCheck{Name: "Inertia frontend", OK: true, Detail: "app.html + Svelte pages + gonertia"}
}

func checkViteConfig(dir string) doctorCheck {
	path := filepath.Join(dir, "vite.config.js")
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{
			Name:    "vite.config.js",
			Detail:  "missing",
			FixHint: "re-run cais new or restore vite.config.js from Cais Inertia scaffold",
		}
	}
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return doctorCheck{Name: "vite.config.js", OK: true, Detail: "present (package.json unreadable)"}
	}
	if !strings.Contains(string(data), "@inertiajs/svelte") {
		return doctorCheck{
			Name:    "vite.config.js",
			Detail:  "package.json missing @inertiajs/svelte",
			FixHint: "cais install",
		}
	}
	buildDir := filepath.Join(dir, "web/static/build")
	if _, err := os.Stat(buildDir); err != nil {
		return doctorCheck{
			Name:     "vite.config.js",
			Optional: true,
			Detail:   "Vite build output missing — run npm run build before deploy",
			FixHint:  "npm run build (or cais dev once)",
		}
	}
	return doctorCheck{Name: "vite.config.js", OK: true, Detail: "Vite + @inertiajs/svelte configured"}
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

var (
	writeTimeoutRe     = regexp.MustCompile(`WriteTimeout:\s*(\d+)\s*\*\s*time\.Second`)
	cacheVersionDoctor = regexp.MustCompile(`const CACHE_VERSION = \d+;`)
)

func checkSSEWriteTimeout(dir string) doctorCheck {
	ssePath := filepath.Join(dir, "web/static/js/sse-ext.min.js")
	if _, err := os.Stat(ssePath); err != nil {
		return doctorCheck{Name: "SSE WriteTimeout", OK: true, Detail: "skipped (no sse-ext.min.js)"}
	}
	appPath := filepath.Join(dir, "internal/app/app.go")
	data, err := os.ReadFile(appPath)
	if err != nil {
		return doctorCheck{Name: "SSE WriteTimeout", OK: true, Detail: "skipped (no internal/app/app.go)"}
	}
	m := writeTimeoutRe.FindStringSubmatch(string(data))
	if m == nil {
		return doctorCheck{Name: "SSE WriteTimeout", OK: true, Detail: "skipped (WriteTimeout not detected)"}
	}
	if m[1] == "0" {
		return doctorCheck{Name: "SSE WriteTimeout", OK: true, Detail: "disabled for streaming"}
	}
	return doctorCheck{
		Name:     "SSE WriteTimeout",
		Optional: true,
		Detail:   fmt.Sprintf("WriteTimeout: %s*time.Second kills long-lived SSE connections", m[1]),
		FixHint:  "set WriteTimeout: 0 in internal/app/app.go (see pkg/cais/stream)",
	}
}

func checkSSEExt(dir string) doctorCheck {
	path := filepath.Join(dir, "web/static/js/sse-ext.min.js")
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{
			Name:    "sse-ext.min.js",
			Detail:  "missing",
			FixHint: "re-run cais new, cais pwa, or copy from Cais web/static/js/sse-ext.min.js",
		}
	}
	return doctorCheck{Name: "sse-ext.min.js", OK: true}
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
	path := filepath.Join(dir, cssOutput)
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{Name: "tailwind css", Detail: "styles.css missing", FixHint: "cais css"}
	}
	if !stylesCSSReady(dir) {
		return doctorCheck{
			Name:    "tailwind css",
			Detail:  "styles.css empty or not built (app will look unstyled)",
			FixHint: "cais css",
		}
	}
	return doctorCheck{Name: "tailwind css", OK: true}
}

// stylesCSSReady reports whether web/static/css/styles.css looks like a real Tailwind build.
// Scaffold writes a comment-only stub; git clones omit the gitignored file entirely (#141).
func stylesCSSReady(dir string) bool {
	body, err := os.ReadFile(filepath.Join(dir, cssOutput))
	if err != nil {
		return false
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return false
	}
	// Comment-only / placeholder stubs ship with cais new until `cais css` runs.
	withoutComments := regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(trimmed, "")
	withoutComments = strings.TrimSpace(withoutComments)
	if withoutComments == "" {
		return false
	}
	// Real Tailwind output always contains a CSS rule (selector + block).
	return strings.Contains(withoutComments, "{")
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
