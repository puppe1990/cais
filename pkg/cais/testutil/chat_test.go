package testutil

import (
	"strings"
	"testing"
)

func TestAssertHTMLContains_passesWhenAllPresent(t *testing.T) {
	AssertHTMLContains(t, `<div id="chat-history">ok</div>`, `id="chat-history"`, "ok")
}

func TestAssertChatMarkers_requiresChatDOM(t *testing.T) {
	body := strings.Join([]string{
		`id="chat-history"`,
		`id="chat-messages"`,
		`data-cais-chat="true"`,
		`data-cais-chat-form`,
	}, "\n")
	AssertChatMarkers(t, body)
}
