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
	RoleDetail    Role = "detail"
)

// MaxMessageChars is the default soft limit for a single chat turn before truncation.
// Protects against huge tool dumps, context pollution and slow/unsafe renders.
const MaxMessageChars = 12000

const assistantBubbleClass = "cais-chat-bubble assistant max-w-[85%] rounded-2xl rounded-bl-sm bg-white border border-slate-200 px-4 py-2 text-sm text-slate-800 shadow-xs"
const userBubbleClass = "cais-chat-bubble user max-w-[85%] rounded-2xl rounded-br-sm bg-indigo-600 px-4 py-2 text-sm text-white shadow-xs"
const detailBubbleClass = "cais-chat-bubble detail max-w-[85%] rounded-xl rounded-bl-sm bg-slate-50 border border-slate-200 px-3 py-2 text-xs text-slate-600 shadow-xs self-start"

// DetailBubble renders collapsible tool/log output without polluting assistant bubbles.
func DetailBubble(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	escaped := html.EscapeString(text)
	return fmt.Sprintf(
		`<details class="%s"><summary class="cursor-pointer font-medium text-slate-500">Details</summary><pre class="mt-1 whitespace-pre-wrap">%s</pre></details>`,
		detailBubbleClass, escaped,
	)
}

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

// Truncate returns a safe-to-display prefix of text.
// Cuts near a word boundary when possible and appends a truncation marker.
// Use for large assistant turns or tool output before putting into bubbles.
func Truncate(text string, max int) string {
	if max <= 0 {
		max = MaxMessageChars
	}
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	// Prefer cutting at whitespace near the limit (avoid mid-word when easy).
	cut := max
	for i := max; i > max-180 && i > 10; i-- {
		if text[i] == ' ' || text[i] == '\n' || text[i] == '\t' {
			cut = i
			break
		}
	}
	return strings.TrimSpace(text[:cut]) + " … [truncated]"
}

// SafeMessageBubble is MessageBubble but truncates huge single messages for perf and safety.
// Apps dealing with agents that can emit megabytes of context should use this (or Truncate first).
func SafeMessageBubble(role Role, text string, at time.Time) string {
	return MessageBubble(role, Truncate(text, 0), at)
}

// TrimForDisplay returns the tail of the slice (most recent items) up to max.
// Keeps full history in DB but renders only a window. Pair with "load older" UI for deep history.
func TrimForDisplay[T any](items []T, max int) []T {
	if max <= 0 || len(items) <= max {
		return items
	}
	return items[len(items)-max:]
}

// SelectWindowWithLastUser returns a display window of at most max items from the recent tail,
// while guaranteeing that the most recent message for which isUser returns true is present.
// This eliminates the common app-level "tail + pin last user message" hacks for polluted histories.
func SelectWindowWithLastUser[T any](items []T, max int, isUser func(T) bool) []T {
	n := len(items)
	if n == 0 || max <= 0 {
		return items
	}
	if n <= max {
		return items
	}

	// locate most recent user
	lastUser := -1
	for i := n - 1; i >= 0; i-- {
		if isUser(items[i]) {
			lastUser = i
			break
		}
	}

	window := TrimForDisplay(items, max)
	if lastUser < 0 {
		return window
	}

	// already inside the tail window?
	windowStart := n - len(window)
	if lastUser >= windowStart {
		return window
	}

	// include it by shifting start back (still aim for ~max size)
	start := lastUser
	if n-start > max {
		start = n - max
	}
	if start < 0 {
		start = 0
	}
	return items[start:]
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
