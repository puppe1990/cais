package cli

import (
	"os"
	"path/filepath"
	"strings"
)

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
	body := string(data)
	if !cacheVersionDoctor.Match(data) {
		return doctorCheck{
			Name:     "PWA cache version",
			Optional: true,
			Detail:   "legacy sw.js without CACHE_VERSION",
			FixHint:  "run cais pwa to refresh assets, then cais pwa --bump before phone testing",
		}
	}
	// SPA + Tailwind use stable paths; cache-first for /static/build keeps stale main.js.
	if !strings.Contains(body, "/static/build/") {
		return doctorCheck{
			Name:     "PWA cache version",
			Optional: true,
			Detail:   "sw.js missing network-first for /static/build/ (stale SPA after vite rebuild)",
			FixHint:  "run cais pwa to refresh sw.js, or set network-first for /static/build/ and /static/css/",
		}
	}
	return doctorCheck{Name: "PWA cache version", OK: true, Detail: "CACHE_VERSION + network-first /static/build — bump after HTML/template changes"}
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
	path := filepath.Join(dir, "web/static/js/cais-core.js")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "SSE reconnect", OK: true, Detail: "skipped (no cais-core.js)"}
	}
	content := string(data)
	if strings.Contains(content, "reconnectChatSSE") && strings.Contains(content, "htmx:sseClose") {
		return doctorCheck{Name: "SSE reconnect", OK: true, Detail: "cais-core.js reconnects SSE after hx-boost"}
	}
	return doctorCheck{
		Name:     "SSE reconnect",
		Optional: true,
		Detail:   "cais-core.js missing hx-boost SSE reconnect helpers",
		FixHint:  "run cais pwa to refresh cais-core.js from framework",
	}
}

func chatUsesAgentSlots(dir string) bool {
	partials := filepath.Join(dir, "web/templates/partials")
	entries, err := os.ReadDir(partials)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(partials, e.Name()))
		if err != nil {
			continue
		}
		content := string(data)
		if strings.Contains(content, `id="chat-stream"`) ||
			strings.Contains(content, `id="chat-live"`) ||
			strings.Contains(content, `data-cais-chat="true"`) {
			return true
		}
	}
	return false
}

func checkChatAgentJS(dir string) doctorCheck {
	if !chatUsesAgentSlots(dir) {
		return doctorCheck{Name: "chat agent JS", OK: true, Detail: "skipped (no agent chat partial)"}
	}
	path := filepath.Join(dir, "web/static/js/cais-chat.js")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "chat agent JS", OK: true, Detail: "skipped (no cais-chat.js)"}
	}
	content := string(data)
	if strings.Contains(content, "finalizeChatStream") && strings.Contains(content, "data-cais-chat") {
		return doctorCheck{Name: "chat agent JS", OK: true, Detail: "cais-chat.js finalizes multi-slot SSE chat"}
	}
	return doctorCheck{
		Name:     "chat agent JS",
		Optional: true,
		Detail:   "agent chat partial present but cais-chat.js missing finalizeChatStream",
		FixHint:  "run cais pwa to refresh cais-chat.js from framework",
	}
}

func checkChatScrollContainer(dir string) doctorCheck {
	if !chatUsesAgentSlots(dir) {
		return doctorCheck{Name: "chat scroll container", OK: true, Detail: "skipped (no agent chat partial)"}
	}
	partials := filepath.Join(dir, "web/templates/partials")
	entries, err := os.ReadDir(partials)
	if err != nil {
		return doctorCheck{Name: "chat scroll container", OK: true, Detail: "skipped (no partials)"}
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(partials, e.Name()))
		if err != nil {
			continue
		}
		content := string(data)
		if !strings.Contains(content, `data-cais-chat="true"`) {
			continue
		}
		if strings.Contains(content, `id="chat-messages"`) {
			if !strings.Contains(content, "overflow-x-hidden") && !strings.Contains(content, "overflow-x: hidden") {
				return doctorCheck{
					Name:     "chat scroll container",
					Optional: true,
					Detail:   "#chat-messages missing overflow-x-hidden (horizontal scroll risk)",
					FixHint:  "add overflow-x-hidden max-w-full to #chat-messages (see chat_sse_agent.html)",
				}
			}
			return doctorCheck{Name: "chat scroll container", OK: true, Detail: "#chat-messages scroll container present"}
		}
		return doctorCheck{
			Name:     "chat scroll container",
			Optional: true,
			Detail:   "data-cais-chat without #chat-messages scroll container",
			FixHint:  "wrap chat slots in #chat-messages with overflow-y-auto (see chat_sse_agent.html)",
		}
	}
	return doctorCheck{Name: "chat scroll container", OK: true, Detail: "skipped (no data-cais-chat)"}
}

func appUsesChatForm(dir string) bool {
	root := filepath.Join(dir, "web", "templates")
	found := false
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		if strings.Contains(content, "hxChatForm") || strings.Contains(content, `data-cais-chat-form`) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func checkChatEnterSubmitJS(dir string) doctorCheck {
	if !appUsesChatForm(dir) {
		return doctorCheck{Name: "chat enter-submit JS", OK: true, Detail: "skipped (no chat form)"}
	}
	path := filepath.Join(dir, "web/static/js/cais-chat.js")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "chat enter-submit JS", OK: true, Detail: "skipped (no cais-chat.js)"}
	}
	content := string(data)
	if strings.Contains(content, "bindChatEnterSubmit") && strings.Contains(content, "data-cais-chat-form") {
		return doctorCheck{Name: "chat enter-submit JS", OK: true, Detail: "cais-chat.js handles Enter-to-send on chat forms"}
	}
	return doctorCheck{
		Name:     "chat enter-submit JS",
		Optional: true,
		Detail:   "chat form present but cais-chat.js missing bindChatEnterSubmit",
		FixHint:  "run cais pwa to refresh cais.js from framework",
	}
}

func checkChatFormCSS(dir string) doctorCheck {
	if !appUsesChatForm(dir) && !chatUsesAgentSlots(dir) {
		return doctorCheck{Name: "chat form CSS", OK: true, Detail: "skipped (no chat UI)"}
	}
	path := filepath.Join(dir, "input.css")
	data, err := os.ReadFile(path)
	if err != nil {
		return doctorCheck{Name: "chat form CSS", OK: true, Detail: "skipped (no input.css)"}
	}
	content := string(data)
	hasShell := strings.Contains(content, ".cais-chat-shell")
	hasSubmit := strings.Contains(content, "form[data-cais-chat-form]")
	if hasShell && hasSubmit {
		return doctorCheck{Name: "chat form CSS", OK: true, Detail: "mobile chat shell + submit indicator CSS present"}
	}
	missing := []string{}
	if !hasShell {
		missing = append(missing, ".cais-chat-shell")
	}
	if !hasSubmit {
		missing = append(missing, "form[data-cais-chat-form]")
	}
	return doctorCheck{
		Name:     "chat form CSS",
		Optional: true,
		Detail:   "input.css missing: " + strings.Join(missing, ", "),
		FixHint:  "run cais css after updating input.css from Cais scaffold",
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
