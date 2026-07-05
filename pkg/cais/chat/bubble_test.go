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

func TestTruncate_shortDoesNothing(t *testing.T) {
	got := Truncate("hello world", 100)
	if got != "hello world" {
		t.Errorf("got %q", got)
	}
}

func TestTruncate_longTruncatesWithMarker(t *testing.T) {
	long := strings.Repeat("x", 20000)
	got := Truncate(long, 100)
	if !strings.HasSuffix(got, "… [truncated]") {
		t.Errorf("expected truncation marker, got %q", got)
	}
	if len(got) > 120 {
		t.Errorf("truncated too long: %d", len(got))
	}
}

func TestTruncate_respectsWordBoundaryWhenPossible(t *testing.T) {
	text := "hello world this is a long sentence with spaces " + strings.Repeat("y", 50)
	got := Truncate(text, 30)
	if !strings.HasSuffix(got, "… [truncated]") {
		t.Errorf("expected ends with truncation marker, got %q", got)
	}
	if len(got) > 50 {
		t.Errorf("truncated result still too long: %d", len(got))
	}
}

func TestSafeMessageBubble_truncatesLargeContent(t *testing.T) {
	huge := strings.Repeat("A", 30000) + " secret"
	got := SafeMessageBubble(RoleAssistant, huge, timeFromTest())
	if strings.Contains(got, "secret") {
		t.Error("should not include content after truncation point")
	}
	if !strings.Contains(got, "[truncated]") {
		t.Error("expected [truncated] marker in safe bubble")
	}
}

func TestMessageBubble_keepsFullContent_unlessUsingSafe(t *testing.T) {
	huge := strings.Repeat("B", 15000)
	got := MessageBubble(RoleUser, huge, timeFromTest())
	if !strings.Contains(got, strings.Repeat("B", 100)) {
		t.Error("MessageBubble should preserve original (caller responsible for size)")
	}
}

func TestTrimForDisplay_keepsRecent(t *testing.T) {
	all := []string{"m1", "m2", "m3", "m4", "m5"}
	got := TrimForDisplay(all, 2)
	if len(got) != 2 || got[0] != "m4" || got[1] != "m5" {
		t.Errorf("got %v, want last 2", got)
	}
}

func TestTrimForDisplay_smallerThanLimitReturnsAll(t *testing.T) {
	all := []int{1, 2, 3}
	if got := TrimForDisplay(all, 10); len(got) != 3 {
		t.Errorf("unexpected trim: %v", got)
	}
}

type testMsg struct {
	Role string
	Body string
}

func isUser(m testMsg) bool { return m.Role == "user" }

func TestSelectWindowWithLastUser_includesPinnedEvenIfOld(t *testing.T) {
	history := []testMsg{
		{Role: "user", Body: "q1"},
		{Role: "assistant", Body: strings.Repeat("x", 1000)},
		{Role: "user", Body: "q2"},
		{Role: "assistant", Body: "a2"},
		{Role: "user", Body: "last question here"},
		{Role: "assistant", Body: "a3"},
		{Role: "assistant", Body: "a4"},
	}
	got := SelectWindowWithLastUser(history, 3, isUser)
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3: %v", len(got), got)
	}
	foundUser := false
	for _, m := range got {
		if m.Body == "last question here" {
			foundUser = true
		}
	}
	if !foundUser {
		t.Error("last user question was not pinned into the window")
	}
}

func TestSelectWindowWithLastUser_noUserJustTrims(t *testing.T) {
	history := []testMsg{{Role: "assistant", Body: "1"}, {Role: "assistant", Body: "2"}, {Role: "assistant", Body: "3"}}
	got := SelectWindowWithLastUser(history, 2, isUser)
	if len(got) != 2 || got[0].Body != "2" {
		t.Errorf("got %v", got)
	}
}

func TestDetailBubbleWithTitle_andToolHelpers(t *testing.T) {
	d := DetailBubbleWithTitle("mytool", "some output\nwith\nlines")
	if !strings.Contains(d, "<summary") || !strings.Contains(d, "mytool") || !strings.Contains(d, "some output") {
		t.Errorf("bad detail with title: %s", d)
	}

	tc := ToolCallBubble("fs.read", `{"path":"x"}`)
	if !strings.Contains(tc, "tool: fs.read") {
		t.Error("ToolCallBubble title wrong")
	}

	tr := ToolResultBubble("42")
	if !strings.Contains(tr, "tool result") {
		t.Error("ToolResultBubble title wrong")
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
