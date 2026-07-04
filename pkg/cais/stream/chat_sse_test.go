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
