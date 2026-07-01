package cli

import (
	"os"
	"strings"
	"testing"
)

func TestCLI_Help_IncludesConsole(t *testing.T) {
	var buf strings.Builder
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "cais console") {
		t.Error("help missing cais console")
	}
}

func TestCLI_Console_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"console"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}
