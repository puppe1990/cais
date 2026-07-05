package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldStreamChat_CreatesFiles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "streamapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "streamapp",
		ModulePath: "github.com/puppe1990/streamapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldStreamChat(appDir, streamOpts{}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/models/conversation.go",
		"internal/models/message.go",
		"internal/handlers/chat.go",
		"internal/handlers/chat_test.go",
		"web/templates/pages/conversations.html",
		"web/templates/pages/chat.html",
		"web/templates/partials/message.html",
		"web/templates/partials/chat_sse_agent.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	storeBody, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(storeBody), "ListConversations") {
		t.Error("store.go missing ListConversations")
	}
	if !strings.Contains(string(storeBody), "InsertMessage") {
		t.Error("store.go missing InsertMessage")
	}

	routesBody, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	routes := string(routesBody)
	for _, want := range []string{
		`NewChatHandler`,
		`/chat/{id}/stream`,
		`cais.IntParam("id", chat.Stream)`,
		`middleware.RequireAuth("/login")`,
	} {
		if !strings.Contains(routes, want) {
			t.Errorf("routes.go missing %q", want)
		}
	}

	chatBody, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/chat.go"))
	if err != nil {
		t.Fatal(err)
	}
	handler := string(chatBody)
	for _, want := range []string{
		`pkg/cais/chat`,
		`chat.WriteStream`,
		`chat.MessageBubble`,
		`chat.WriteMessage`,
	} {
		if !strings.Contains(handler, want) {
			t.Errorf("chat.go missing %q", want)
		}
	}

	testBody, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/chat_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	tests := string(testBody)
	for _, want := range []string{
		`testutil.AssertChatMarkers`,
		`TestChatHandler_Show_NotFound_Returns404`,
		`TestChatHandler_PostMessage_ReturnsUserBubble`,
	} {
		if !strings.Contains(tests, want) {
			t.Errorf("chat_test.go missing %q", want)
		}
	}
}

func TestCLI_GenerateStreamChatDryRun(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "streamdry")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "streamdry",
		ModulePath: "github.com/puppe1990/streamdry",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	c := &CLI{Out: discardWriter{}}
	if err := c.Run([]string{"g", "--dry-run", "stream", "chat"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(appDir, "internal/handlers/chat.go")); !os.IsNotExist(err) {
		t.Error("dry-run should not create chat.go")
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
