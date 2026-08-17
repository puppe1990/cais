package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile_missingIsNoop(t *testing.T) {
	if err := LoadFile(filepath.Join(t.TempDir(), ".env")); err != nil {
		t.Fatalf("missing file should be no-op, got %v", err)
	}
}

func TestLoadFile_setsUnsetKeysOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("DOTENV_NEW=from-file\nDOTENV_EXISTING=from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DOTENV_EXISTING", "from-process")
	os.Unsetenv("DOTENV_NEW")
	t.Cleanup(func() { os.Unsetenv("DOTENV_NEW") })

	if err := LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("DOTENV_NEW"); got != "from-file" {
		t.Fatalf("DOTENV_NEW = %q, want from-file", got)
	}
	if got := os.Getenv("DOTENV_EXISTING"); got != "from-process" {
		t.Fatalf("DOTENV_EXISTING = %q, want from-process", got)
	}
}

func TestLoadFile_emptyProcessEnvWins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("DOTENV_EMPTY=from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOTENV_EMPTY", "")

	if err := LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("DOTENV_EMPTY"); got != "" {
		t.Fatalf("empty process env should win, got %q", got)
	}
}
