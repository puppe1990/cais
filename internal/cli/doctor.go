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
		checkHTMX(dir),
		checkSSEExt(dir),
		checkSSEWriteTimeout(dir),
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
	if opts.Mobile {
		checks = append(checks,
			checkFlashTemplate(dir),
			checkGoogleFonts(dir),
			checkPWACacheVersion(dir),
			checkChatSSEPattern(dir),
			checkSSEReconnectJS(dir),
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

func checkFlashTemplate(dir string) doctorCheck {
	path := filepath.Join(dir, "web/templates/layouts/base.html")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "flash template", OK: true, Detail: "skipped (no base.html)"}
	}
	content := string(data)
	if strings.Contains(content, "flashMessage") || strings.Contains(content, ".Flash.Message") {
		return doctorCheck{Name: "flash template", OK: true, Detail: "uses flashMessage or .Flash.Message"}
	}
	if strings.Contains(content, "{{ .Flash }}") || strings.Contains(content, "{{.Flash}}") {
		return doctorCheck{
			Name:     "flash template",
			Optional: true,
			Detail:   "renders struct as {notice text} — use {{ flashMessage .Flash }}",
			FixHint:  "replace {{ .Flash }} with {{ flashMessage .Flash }} in layouts/base.html",
		}
	}
	return doctorCheck{Name: "flash template", OK: true, Detail: "no flash markup detected"}
}

func checkGoogleFonts(dir string) doctorCheck {
	path := filepath.Join(dir, "input.css")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "CSP fonts", OK: true, Detail: "skipped (no input.css)"}
	}
	if strings.Contains(string(data), "fonts.googleapis.com") {
		return doctorCheck{
			Name:     "CSP fonts",
			Optional: true,
			Detail:   "Google Fonts @import blocked by default CSP (style-src 'self')",
			FixHint:  "remove fonts.googleapis.com from input.css; use system font stack in tailwind.config.js",
		}
	}
	return doctorCheck{Name: "CSP fonts", OK: true, Detail: "no external font imports"}
}

func checkPWACacheVersion(dir string) doctorCheck {
	path := filepath.Join(dir, "web/static/js/sw.js")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "PWA cache version", OK: true, Detail: "skipped (no sw.js)"}
	}
	if cacheVersionDoctor.Match(data) {
		return doctorCheck{Name: "PWA cache version", OK: true, Detail: "CACHE_VERSION present — run cais pwa --bump after template changes"}
	}
	return doctorCheck{
		Name:     "PWA cache version",
		Optional: true,
		Detail:   "legacy sw.js without CACHE_VERSION",
		FixHint:  "run cais pwa to refresh assets, then cais pwa --bump before phone testing",
	}
}

func checkChatSSEPattern(dir string) doctorCheck {
	path := filepath.Join(dir, "web/templates/partials/chat_sse.html")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "chat SSE pattern", OK: true, Detail: "skipped (no chat_sse.html)"}
	}
	content := string(data)
	missing := []string{}
	for _, want := range []string{`id="chat-history"`, `id="chat-sse"`, `hx-swap="beforeend"`, `data-cais-sse-persist`} {
		if !strings.Contains(content, want) {
			missing = append(missing, want)
		}
	}
	if len(missing) == 0 {
		return doctorCheck{Name: "chat SSE pattern", OK: true, Detail: "append-only SSE partial present"}
	}
	return doctorCheck{
		Name:     "chat SSE pattern",
		Optional: true,
		Detail:   "chat_sse.html missing: " + strings.Join(missing, ", "),
		FixHint:  "run cais pwa or copy chat_sse.html from Cais scaffold",
	}
}

func checkSSEReconnectJS(dir string) doctorCheck {
	path := filepath.Join(dir, "web/static/js/cais.js")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "SSE reconnect", OK: true, Detail: "skipped (no cais.js)"}
	}
	content := string(data)
	if strings.Contains(content, "reconnectChatSSE") && strings.Contains(content, "htmx:sseClose") {
		return doctorCheck{Name: "SSE reconnect", OK: true, Detail: "cais.js reconnects SSE after hx-boost"}
	}
	return doctorCheck{
		Name:     "SSE reconnect",
		Optional: true,
		Detail:   "cais.js missing hx-boost SSE reconnect helpers",
		FixHint:  "run cais pwa to refresh cais.js from framework",
	}
}

func checkHealthLANURLs(dir string) doctorCheck {
	path := filepath.Join(dir, "internal/app/app.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "health lan_urls", OK: true, Detail: "skipped (no app.go)"}
	}
	content := string(data)
	if strings.Contains(content, "http://http://") {
		return doctorCheck{
			Name:     "health lan_urls",
			Optional: true,
			Detail:   "malformed double http:// in health handler",
			FixHint:  "use netutil.HealthPayload(status, cfg.Port) — never concatenate APP_URL + port manually",
		}
	}
	if strings.Contains(content, "netutil.HealthPayload") {
		return doctorCheck{Name: "health lan_urls", OK: true, Detail: "uses netutil.HealthPayload"}
	}
	return doctorCheck{
		Name:     "health lan_urls",
		Optional: true,
		Detail:   "health handler does not expose lan_urls via netutil",
		FixHint:  "use netutil.HealthPayload in healthHandler for phone testing",
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
