package chat

import (
	"fmt"
	"html"
	"strings"
	"time"
)

// Role identifies who sent a chat message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

const assistantBubbleClass = "cais-chat-bubble assistant max-w-[85%] rounded-2xl rounded-bl-sm bg-white border border-slate-200 px-4 py-2 text-sm text-slate-800 shadow-xs"
const userBubbleClass = "cais-chat-bubble user max-w-[85%] rounded-2xl rounded-br-sm bg-indigo-600 px-4 py-2 text-sm text-white shadow-xs"

// LiveBubble is a single assistant fragment updated in #chat-live via event: stream.
func LiveBubble(text string) string {
	return fmt.Sprintf(
		`<div data-cais-live="true" class="%s">%s</div>`,
		assistantBubbleClass,
		html.EscapeString(text),
	)
}

// IsLiveHTML reports whether an SSE HTML fragment targets the live stream slot.
func IsLiveHTML(fragment string) bool {
	return strings.Contains(fragment, `data-cais-live="true"`)
}

// MessageBubble is a persisted row with a UTC datetime for client-side local formatting.
func MessageBubble(role Role, text string, at time.Time) string {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	dt := html.EscapeString(at.UTC().Format(time.RFC3339))
	escaped := html.EscapeString(text)
	switch role {
	case RoleUser:
		return fmt.Sprintf(
			`<div class="cais-msg cais-msg-user max-w-[85%%] ml-auto flex flex-col items-end gap-0.5"><time datetime="%s" class="cais-msg-time"></time><div class="%s">%s</div></div>`,
			dt, userBubbleClass, escaped,
		)
	default:
		return fmt.Sprintf(
			`<div class="cais-msg cais-msg-assistant max-w-[85%%] flex flex-col items-start gap-0.5"><time datetime="%s" class="cais-msg-time"></time><div class="%s">%s</div></div>`,
			dt, assistantBubbleClass, escaped,
		)
	}
}

// ThinkingHTML shows the live thinking indicator (HTMX OOB-friendly).
func ThinkingHTML(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		label = "…"
	}
	if len(label) > 120 {
		label = label[:117] + "…"
	}
	escaped := html.EscapeString(label)
	return fmt.Sprintf(
		`<div id="chat-thinking" hx-swap-oob="true" class="cais-thinking flex items-center gap-2.5 max-w-[85%%] rounded-2xl rounded-bl-sm bg-slate-100 border border-slate-200 px-4 py-3 text-sm text-slate-600 shadow-xs self-start" role="status" aria-live="polite">`+
			`<span class="cais-thinking-dots shrink-0" aria-hidden="true"><span></span><span></span><span></span></span>`+
			`<span id="chat-thinking-label">%s</span></div>`,
		escaped,
	)
}

// ThinkingHiddenHTML hides the thinking indicator via HTMX OOB swap.
func ThinkingHiddenHTML() string {
	return `<div id="chat-thinking" hx-swap-oob="true" class="hidden" aria-hidden="true" role="status"></div>`
}

// IsThinkingHTML reports whether an SSE HTML fragment updates the thinking indicator.
func IsThinkingHTML(fragment string) bool {
	return strings.Contains(fragment, `id="chat-thinking"`)
}
