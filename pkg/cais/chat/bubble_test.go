package chat

import (
	"strings"
	"testing"
	"time"
)

func TestLiveBubble_marksLiveAndEscapes(t *testing.T) {
	got := LiveBubble(`<script>alert("x")</script>`)
	if !strings.Contains(got, `data-cais-live="true"`) {
		t.Error("missing data-cais-live marker")
	}
	if strings.Contains(got, "<script>") {
		t.Errorf("expected escaped HTML, got %q", got)
	}
	if !strings.Contains(got, "cais-chat-bubble") {
		t.Error("missing cais-chat-bubble class")
	}
}

func TestIsLiveHTML(t *testing.T) {
	if !IsLiveHTML(LiveBubble("hi")) {
		t.Error("LiveBubble should be detected as live HTML")
	}
	if IsLiveHTML(`<div>plain</div>`) {
		t.Error("plain div should not be live HTML")
	}
}

func TestMessageBubble_assistantTimestampUTC(t *testing.T) {
	at := time.Date(2026, 7, 4, 21, 30, 0, 0, time.UTC)
	got := MessageBubble(RoleAssistant, "Hello", at)
	for _, want := range []string{
		`datetime="2026-07-04T21:30:00Z"`,
		`class="cais-msg-time"`,
		"cais-msg-assistant",
		"cais-chat-bubble",
		"Hello",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestMessageBubble_userAlignsEnd(t *testing.T) {
	got := MessageBubble(RoleUser, "Hi", time.Now().UTC())
	if !strings.Contains(got, "cais-msg-user") {
		t.Error("user bubble should use cais-msg-user")
	}
	if !strings.Contains(got, "ml-auto") {
		t.Error("user bubble should align end")
	}
}

func TestMessageBubble_zeroTimeUsesNowUTC(t *testing.T) {
	got := MessageBubble(RoleAssistant, "x", time.Time{})
	if !strings.Contains(got, `datetime="`) {
		t.Error("zero time should still emit datetime attribute")
	}
}

func TestThinkingHTML_showsLabelAndEscapes(t *testing.T) {
	got := ThinkingHTML(`<b>wait</b>`)
	for _, want := range []string{
		`id="chat-thinking"`,
		"cais-thinking",
		"cais-thinking-dots",
		`id="chat-thinking-label"`,
		"&lt;b&gt;wait&lt;/b&gt;",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestThinkingHTML_emptyLabelDefaults(t *testing.T) {
	got := ThinkingHTML("   ")
	if !strings.Contains(got, ">…<") {
		t.Errorf("empty label should default to ellipsis, got %q", got)
	}
}

func TestThinkingHiddenHTML(t *testing.T) {
	got := ThinkingHiddenHTML()
	if !strings.Contains(got, `id="chat-thinking"`) || !strings.Contains(got, "hidden") {
		t.Errorf("unexpected hidden thinking HTML: %q", got)
	}
}

func TestIsThinkingHTML(t *testing.T) {
	if !IsThinkingHTML(ThinkingHTML("go")) {
		t.Error("ThinkingHTML should be detected")
	}
	if IsThinkingHTML(LiveBubble("x")) {
		t.Error("live bubble is not thinking HTML")
	}
}

func TestDetailBubble_escapesAndUsesDetailRole(t *testing.T) {
	got := DetailBubble("line1\n<script>")
	for _, want := range []string{
		"cais-chat-bubble detail",
		"<details",
		"<summary",
		"line1",
		"&lt;script&gt;",
		"self-start",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "<script>") {
		t.Errorf("expected escaped HTML, got %q", got)
	}
}

func TestDetailBubble_emptyReturnsEmpty(t *testing.T) {
	if DetailBubble("   ") != "" {
		t.Error("empty detail should return empty string")
	}
}

func TestUnsafeLiveHTML_doesNotEscape(t *testing.T) {
	got := UnsafeLiveHTML(`<strong>bold</strong> <em>from markdown</em>`)
	if !strings.Contains(got, `data-cais-live="true"`) {
		t.Error("must keep live marker")
	}
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Error("must preserve raw HTML for live preview")
	}
	if strings.Contains(got, "&lt;strong") {
		t.Error("must not escape")
	}
}

func TestUnsafeMessageHTML_preservesRendered(t *testing.T) {
	got := UnsafeMessageHTML(RoleAssistant, `<h3>Title</h3><p>para</p>`, timeFromTest())
	if !strings.Contains(got, `<h3>Title</h3>`) {
		t.Error("raw HTML not preserved")
	}
}
