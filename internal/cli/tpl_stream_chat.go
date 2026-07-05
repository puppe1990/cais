package cli

const tplConversationModel = `package models

import "time"

type Conversation struct {
	ID        int64
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
`

const tplMessageModel = `package models

import "time"

type Message struct {
	ID             int64
	ConversationID int64
	Role           string
	Content        string
	CreatedAt      time.Time
}
`

const tplChatHandler = `package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"{{.ModulePath}}/internal/models"
	"{{.ModulePath}}/internal/store"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/stream"
	"github.com/puppe1990/cais/pkg/cais/validate"
)

type ChatHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	catalog  *i18n.Catalog
	cfg      cais.Config
}

func NewChatHandler(renderer *cais.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *ChatHandler {
	return &ChatHandler{renderer: renderer, store: s, site: site, catalog: catalog, cfg: cfg}
}

type conversationsPageData struct {
	meta.Site
	Conversations []models.Conversation
}

type chatPageData struct {
	meta.Site
	Conversation models.Conversation
	Messages     []models.Message
	StreamURL    string
	MessagesURL  string
}

type messagePartialData struct {
	Role    string
	Content string
}

func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	convs, err := h.store.ListConversations()
	if err != nil {
		http.Error(w, "could not load conversations", http.StatusInternalServerError)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "conversations", conversationsPageData{
		Site:          meta.ForRequest(h.site, r),
		Conversations: convs,
	}, h.cfg)
}

func (h *ChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	id, err := h.store.InsertConversation("New chat")
	if err != nil {
		http.Error(w, "could not create conversation", http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, fmt.Sprintf("/chat/%d", id))
}

func (h *ChatHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	conv, err := h.store.FindConversationByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	msgs, err := h.store.ListMessages(id)
	if err != nil {
		http.Error(w, "could not load messages", http.StatusInternalServerError)
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "chat", chatPageData{
		Site:         meta.ForRequest(h.site, r),
		Conversation: conv,
		Messages:     msgs,
		StreamURL:    fmt.Sprintf("/chat/%d/stream", id),
		MessagesURL:  fmt.Sprintf("/chat/%d/messages", id),
	}, h.cfg)
}

func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request, id int64) {
	stream.RelaySSE(w)
	lastID := int64(0)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			msgs, err := h.store.ListMessagesSince(id, lastID)
			if err != nil {
				continue
			}
			for _, m := range msgs {
				var buf bytes.Buffer
				if err := h.renderer.RenderPartial(&buf, "message", messagePartialData{Role: m.Role, Content: m.Content}); err != nil {
					continue
				}
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", buf.String())
				_ = stream.Flush(w)
				lastID = m.ID
			}
		}
	}
}

func (h *ChatHandler) PostMessage(w http.ResponseWriter, r *http.Request, id int64) {
	content := r.FormValue("content")
	var errs validate.FieldErrors
	if content == "" {
		errs.Add("content", "Message is required")
	}
	if errs.Any() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if _, err := h.store.FindConversationByID(id); err != nil {
		http.NotFound(w, r)
		return
	}
	if _, err := h.store.InsertMessage(models.Message{ConversationID: id, Role: "user", Content: content}); err != nil {
		http.Error(w, "could not save message", http.StatusInternalServerError)
		return
	}
	_ = h.store.UpdateConversationTitle(id, content)
	// Demo echo — replace with your agent bridge (OpenCode, Grok CLI, etc.).
	if _, err := h.store.InsertMessage(models.Message{ConversationID: id, Role: "assistant", Content: "Echo: " + content}); err != nil {
		http.Error(w, "could not save reply", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := h.renderer.RenderPartial(&buf, "message", messagePartialData{Role: "user", Content: content}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request, id int64) {
	msgs, err := h.store.ListMessages(id)
	if err != nil {
		http.Error(w, "could not load messages", http.StatusInternalServerError)
		return
	}
	type row struct {
		Role    string
		Content string
	}
	rows := make([]row, len(msgs))
	for i, m := range msgs {
		rows[i] = row{Role: m.Role, Content: m.Content}
	}
	if err := h.renderer.RenderPartial(w, "chat_history", struct{ Messages []row }{Messages: rows}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
`

const tplChatHandlerTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/testutil"
)

func TestChatHandler_List_Returns200(t *testing.T) {
	h := NewChatHandler(setupTestRenderer(t), setupTestStore(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/chat", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestChatHandler_Show_Returns200(t *testing.T) {
	s := setupTestStore(t)
	id, err := s.InsertConversation("Test")
	if err != nil {
		t.Fatal(err)
	}
	h := NewChatHandler(setupTestRenderer(t), s, testSite(), testCatalog(), cais.Config{})

	req := testutil.NewRequest(http.MethodGet, "/chat/1", testutil.PathValue("id", "1"))
	rr := httptest.NewRecorder()
	h.Show(rr, req, id)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "chat-history") {
		t.Error("body missing chat-history")
	}
}
`

const tplConversationsPage = `{{"{{"}} define "conversations" {{"}}"}}
<section class="max-w-2xl mx-auto px-4 py-8">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold text-slate-900">Conversations</h1>
    <form method="post" action="/chat">
      {{"{{"}} csrfField .CSRFToken {{"}}"}}
      <button type="submit" class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-bold text-white">New chat</button>
    </form>
  </div>
  <ul class="space-y-2">
    {{"{{"}} range .Conversations {{"}}"}}
    <li>
      <a href="/chat/{{"{{"}} .ID {{"}}"}}" class="block rounded-xl border border-slate-200 bg-white px-4 py-3 hover:border-indigo-300">
        <span class="font-medium text-slate-900">{{"{{"}} if .Title {{"}}"}}{{"{{"}} .Title {{"}}"}}{{"{{"}} else {{"}}"}}Untitled{{"{{"}} end {{"}}"}}</span>
      </a>
    </li>
    {{"{{"}} else {{"}}"}}
    <li class="text-sm text-slate-500">No conversations yet.</li>
    {{"{{"}} end {{"}}"}}
  </ul>
</section>
{{"{{"}} end {{"}}"}}
`

const tplChatPage = `{{"{{"}} define "chat" {{"}}"}}
<section class="max-w-2xl mx-auto px-4 py-6 flex flex-col min-h-[70vh]">
  <div class="flex items-center gap-3 mb-4">
    <a href="/chat" class="text-slate-500 hover:text-slate-900" aria-label="Back to conversations">←</a>
    <h1 class="text-lg font-bold text-slate-900 truncate">{{"{{"}} if .Conversation.Title {{"}}"}}{{"{{"}} .Conversation.Title {{"}}"}}{{"{{"}} else {{"}}"}}Chat{{"{{"}} end {{"}}"}}</h1>
  </div>
  {{"{{"}} template "chat_sse_agent" . {{"}}"}}
  <form class="mt-4 flex gap-2" {{"{{"}} hxChatForm (printf "/chat/%d/messages" .Conversation.ID) "#chat-thinking" {{"}}"}}>
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <textarea name="content" rows="2" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Message…" required></textarea>
    <button type="submit" class="rounded-xl bg-indigo-600 px-4 py-2 text-sm font-bold text-white self-end">Send</button>
  </form>
</section>
{{"{{"}} end {{"}}"}}
`

const tplMessagePartial = `{{"{{"}} define "message" {{"}}"}}
<div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
  <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplChatHistoryPartial = `{{"{{"}} define "chat_history" {{"}}"}}
{{"{{"}} range .Messages {{"}}"}}
<div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
  <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
{{"{{"}} end {{"}}"}}
`

const tplChatMigration = `-- up
CREATE TABLE conversations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);

-- down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
`
