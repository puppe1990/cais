package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var goModReplaceLine = regexp.MustCompile(`(?m)^replace ` + regexp.QuoteMeta(frameworkModule) + ` => .*\n?`)

// setGoModReplace adds or updates the local framework replace directive in go.mod.
func setGoModReplace(appDir, replacePath string) error {
	replacePath = strings.TrimSpace(replacePath)
	if replacePath == "" {
		return fmt.Errorf("replace path is required")
	}
	path := filepath.Join(appDir, "go.mod")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	line := fmt.Sprintf("replace %s => %s\n", frameworkModule, replacePath)
	if goModReplaceLine.MatchString(content) {
		content = goModReplaceLine.ReplaceAllString(content, line)
	} else {
		content = strings.TrimRight(content, "\n") + "\n\n" + line
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func removeGoModReplace(appDir string) error {
	path := filepath.Join(appDir, "go.mod")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := goModReplaceLine.ReplaceAllString(string(body), "")
	content = strings.TrimRight(content, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

func (c *CLI) cmdLink(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}

	if len(args) > 0 && args[0] == "--unlink" {
		if err := removeGoModReplace(dir); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(c.Out, "→ removed local cais replace from go.mod")
		return runCmd(dir, "go", "mod", "tidy")
	}

	replacePath := ""
	if len(args) > 0 {
		replacePath = args[0]
	} else {
		replacePath = findLocalCaisReplace(dir)
	}
	if replacePath == "" {
		return fmt.Errorf("could not find local Cais checkout — pass path: cais link ../Cais")
	}

	if err := setGoModReplace(dir, replacePath); err != nil {
		return err
	}
	printLinkMessage(c.Out, replacePath)
	return runCmd(dir, "go", "mod", "tidy")
}

func printLinkMessage(w io.Writer, replacePath string) {
	_, _ = fmt.Fprintf(w, "→ linked %s => %s\n", frameworkModule, replacePath)
	_, _ = fmt.Fprintln(w, "  run cais test to verify; cais link --unlink to restore remote module")
}
