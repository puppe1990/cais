package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI_Help(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "cais new") {
		t.Error("help missing cais new")
	}
}

func TestCLI_Help_IncludesResource(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "[--dry-run] resource") {
		t.Error("help missing g resource")
	}
}

func TestCLI_Help_IncludesModuleFlag(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "--module") {
		t.Error("help missing --module flag")
	}
}
