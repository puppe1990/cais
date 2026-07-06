package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func patchStoreForStreamChat(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "ListConversations") {
		return nil
	}

	ifaceMarker := "\n\tClose() error"
	if !strings.Contains(content, ifaceMarker) {
		return fmt.Errorf("could not patch store interface")
	}
	ifaceInsert := `
	ListConversations() ([]models.Conversation, error)
	FindConversationByID(id int64) (models.Conversation, error)
	InsertConversation(title string) (int64, error)
	UpdateConversationTitle(id int64, title string) error
	ListMessages(conversationID int64) ([]models.Message, error)
	ListMessagesSince(conversationID, afterID int64) ([]models.Message, error)
	InsertMessage(msg models.Message) (int64, error)`
	content = strings.Replace(content, ifaceMarker, ifaceInsert+ifaceMarker, 1)

	implMarker := "\nfunc (s *SQLiteStore) Close()"
	implInsert := streamChatStoreMethods()
	content = strings.Replace(content, implMarker, implInsert+implMarker, 1)

	if !strings.Contains(content, data.ModulePath+"/internal/models") {
		content = strings.Replace(content,
			`_ "modernc.org/sqlite"`,
			`"`+data.ModulePath+`/internal/models"
	_ "modernc.org/sqlite"`,
			1,
		)
	}

	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func streamChatStoreMethods() string {
	return `
func (s *SQLiteStore) ListConversations() ([]models.Conversation, error) {
	rows, err := s.db.Query(` + "`SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC`" + `)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()
	var out []models.Conversation
	for rows.Next() {
		var c models.Conversation
		var created, updated string
		if err := rows.Scan(&c.ID, &c.Title, &created, &updated); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		c.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) FindConversationByID(id int64) (models.Conversation, error) {
	var c models.Conversation
	var created, updated string
	err := s.db.QueryRow(
		` + "`SELECT id, title, created_at, updated_at FROM conversations WHERE id = ?`" + `, id,
	).Scan(&c.ID, &c.Title, &created, &updated)
	if err != nil {
		return c, fmt.Errorf("find conversation: %w", err)
	}
	c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	c.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
	return c, nil
}

func (s *SQLiteStore) InsertConversation(title string) (int64, error) {
	res, err := s.db.Exec(` + "`INSERT INTO conversations (title) VALUES (?)`" + `, title)
	if err != nil {
		return 0, fmt.Errorf("insert conversation: %w", err)
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) UpdateConversationTitle(id int64, title string) error {
	if len(title) > 80 {
		title = title[:80]
	}
	_, err := s.db.Exec(` + "`UPDATE conversations SET title = ?, updated_at = datetime('now') WHERE id = ?`" + `, title, id)
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListMessages(conversationID int64) ([]models.Message, error) {
	rows, err := s.db.Query(
		` + "`SELECT id, conversation_id, role, content, created_at FROM messages WHERE conversation_id = ? ORDER BY id`" + `,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()
	var out []models.Message
	for rows.Next() {
		var m models.Message
		var created string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &created); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) ListMessagesSince(conversationID, afterID int64) ([]models.Message, error) {
	rows, err := s.db.Query(
		` + "`SELECT id, conversation_id, role, content, created_at FROM messages WHERE conversation_id = ? AND id > ? ORDER BY id`" + `,
		conversationID, afterID,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages since: %w", err)
	}
	defer rows.Close()
	var out []models.Message
	for rows.Next() {
		var m models.Message
		var created string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &created); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) InsertMessage(msg models.Message) (int64, error) {
	res, err := s.db.Exec(
		` + "`INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)`" + `,
		msg.ConversationID, msg.Role, msg.Content,
	)
	if err != nil {
		return 0, fmt.Errorf("insert message: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	_, _ = s.db.Exec(` + "`UPDATE conversations SET updated_at = datetime('now') WHERE id = ?`" + `, msg.ConversationID)
	return id, nil
}
`
}

func patchRoutesForStreamChat(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "NewChatHandler") {
		return nil
	}

	insert := `
	chat := handlers.NewChatHandler(deps.Renderer, deps.Store, deps.Site, deps.Catalog, cfg)
	r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
		g.Get("/chat", chat.List)
		g.Post("/chat", chat.Create)
		g.Get("/chat/{id}", cais.IntParam("id", chat.Show))
		g.Get("/chat/{id}/stream", cais.IntParam("id", chat.Stream))
		g.Post("/chat/{id}/messages", cais.IntParam("id", chat.PostMessage))
		g.Get("/chat/{id}/messages", cais.IntParam("id", chat.ListMessages))
	})
`

	updated, err := insertBeforeFunctionEnd(content, "registerRoutes", insert)
	if err != nil {
		return fmt.Errorf("could not patch routes.go: %w", err)
	}
	return updateScaffoldFile(path, []byte(updated), "internal/app/routes.go", dryRun)
}

func patchLayoutNavForStreamChat(dir string, dryRun bool) error {
	path, inertia := layoutNavFile(dir)
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(body)
	if strings.Contains(content, `href="/chat"`) {
		return nil
	}
	marker := layoutNavMarker
	if !strings.Contains(content, marker) {
		marker = "</nav>"
	}
	var navLink string
	if inertia {
		navLink = `    <a href="/chat" use:inertia class="underline">Chat</a>
`
	} else {
		navLink = `          <a href="/chat" data-cais-nav="/chat" hx-boost="true" hx-target="#cais-main" hx-select="#cais-main" hx-push-url="true" hx-swap="morph:innerHTML" class="px-3 py-1.5 rounded-lg text-xs font-bold text-slate-600 hover:text-slate-900 hover:bg-slate-100">Chat</a>
`
	}
	content = strings.Replace(content, marker, navLink+marker, 1)
	rel := strings.TrimPrefix(path, dir+string(os.PathSeparator))
	return updateScaffoldFile(path, []byte(content), rel, dryRun)
}
