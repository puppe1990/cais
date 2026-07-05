package stream

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/testutil"
)

type chatMessage struct {
	Role    string
	Content string
}

func TestChatSSEPartial_appendPattern(t *testing.T) {
	root := testutil.ProjectRoot(t)
	r, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), nil)
	if err != nil {
		t.Fatal(err)
	}

	data := struct {
		StreamURL   string
		MessagesURL string
		Messages    []chatMessage
	}{
		StreamURL: "/chat/1/stream",
		Messages: []chatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	rr := httptest.NewRecorder()
	if err := r.RenderPartial(rr, "chat_sse", data); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()

	if !strings.Contains(body, `id="chat-history"`) {
		t.Error("missing chat-history container")
	}
	if !strings.Contains(body, `id="chat-sse"`) {
		t.Error("missing chat-sse listener")
	}
	if !strings.Contains(body, `sse-connect="/chat/1/stream"`) {
		t.Error("missing sse-connect URL")
	}
	if !strings.Contains(body, `hx-target="#chat-history"`) {
		t.Error("missing hx-target for append pattern")
	}
	if !strings.Contains(body, `hx-swap="beforeend"`) {
		t.Error("missing beforeend swap — SSE must append, not replace history")
	}
	if strings.Contains(body, `sse-swap="innerHTML"`) {
		t.Error("innerHTML sse-swap would wipe chat history")
	}
	if !strings.Contains(body, `id="chat-thinking"`) {
		t.Error("missing chat-thinking indicator")
	}
	if !strings.Contains(body, `data-cais-sse-persist="true"`) {
		t.Error("missing data-cais-sse-persist for hx-boost reconnect")
	}
}

func TestChatSSEAgentPartial_multiSlotPattern(t *testing.T) {
	root := testutil.ProjectRoot(t)
	r, err := cais.NewRendererFromDir(filepath.Join(root, "web", "templates"), nil)
	if err != nil {
		t.Fatal(err)
	}

	data := struct {
		StreamURL   string
		MessagesURL string
		Messages    []chatMessage
	}{
		StreamURL:   "/chat/1/stream",
		MessagesURL: "/chat/1/messages",
		Messages: []chatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	rr := httptest.NewRecorder()
	if err := r.RenderPartial(rr, "chat_sse_agent", data); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()

	for _, want := range []string{
		`data-cais-chat="true"`,
		`id="chat-messages"`,
		`id="chat-history"`,
		`id="chat-stream"`,
		`id="chat-live"`,
		`id="chat-thinking"`,
		`id="chat-sse"`,
		`id="chat-scroll-down"`,
		`sse-swap="stream"`,
		`hx-target="#chat-live"`,
		`hx-swap="innerHTML"`,
		`sse-swap="message"`,
		`hx-target="#chat-stream"`,
		`hx-swap="beforeend"`,
		`sse-swap="thinking"`,
		`data-cais-sse-persist="true"`,
		`data-cais-poll-url="/chat/1/messages"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in agent partial", want)
		}
	}

	// Base render must still produce the contract slots.
	// The partial now supports extension via:
	//   {{ block "chat_extra_stream_slots" . }} ... {{ end }}
	// Consumers define the block to add more sse-swap listeners (e.g. detail, user)
	// without copying the entire chat_sse_agent.html.
}
