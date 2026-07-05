package testutil

import (
	"strings"
	"testing"
)

// AssertHTMLContains fails when body does not contain every substring.
func AssertHTMLContains(t testing.TB, body string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if !strings.Contains(body, sub) {
			t.Errorf("body missing %q", sub)
		}
	}
}

// AssertChatMarkers checks agent-mode chat page/partial DOM hooks.
func AssertChatMarkers(t testing.TB, body string) {
	t.Helper()
	AssertHTMLContains(t, body,
		`id="chat-history"`,
		`id="chat-messages"`,
		`data-cais-chat="true"`,
		`data-cais-chat-form`,
	)
}
