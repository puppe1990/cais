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

// DetailBubbleWithTitle is DetailBubble but allows a custom <summary> label (useful for tool names).
func DetailBubbleWithTitle(title, text string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Details"
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	return fmt.Sprintf(
		`<details class="%s"><summary class="cursor-pointer font-medium text-slate-500">%s</summary><pre class="mt-1 whitespace-pre-wrap">%s</pre></details>`,
		detailBubbleClass, html.EscapeString(title), html.EscapeString(text),
	)
}

// ToolCallBubble renders a collapsible tool invocation (for permissions/audit in agent UIs).
func ToolCallBubble(tool, input string) string {
	if tool == "" {
		tool = "tool"
	}
	return DetailBubbleWithTitle("tool: "+tool, input)
}

// ToolResultBubble renders tool output safely (distinguishes from assistant text).
func ToolResultBubble(output string) string {
	return DetailBubbleWithTitle("tool result", output)
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

// UnsafeLiveHTML returns a live-update fragment containing pre-rendered HTML (e.g. progressive Markdown from Goldmark).
// The caller MUST sanitize/escape untrusted content. No HTML escaping is performed here.
// This enables first-class rich streaming UIs without forcing the app to duplicate the live bubble wrapper.
func UnsafeLiveHTML(htmlContent string) string {
	return fmt.Sprintf(
		`<div data-cais-live="true" class="%s">%s</div>`,
		assistantBubbleClass, htmlContent,
	)
}

// UnsafeMessageHTML is the finalized version of UnsafeLiveHTML for pre-rendered assistant (or user) content.
// Allows apps to do their own rich rendering (markdown, media refs, diffs) for both live preview and final bubbles.
func UnsafeMessageHTML(role Role, htmlContent string, at time.Time) string {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	dt := html.EscapeString(at.UTC().Format(time.RFC3339))
	cls := assistantBubbleClass
	align := "flex flex-col items-start gap-0.5"
	if role == RoleUser {
		cls = userBubbleClass
		align = "flex flex-col items-end gap-0.5 ml-auto"
	}
	return fmt.Sprintf(
		`<div class="cais-msg cais-msg-%s max-w-[85%%] %s"><time datetime="%s" class="cais-msg-time"></time><div class="%s">%s</div></div>`,
		role, align, dt, cls, htmlContent,
	)
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
