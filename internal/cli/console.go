package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func (c *CLI) cmdConsole() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}

	consoleMain := filepath.Join(dir, "cmd", "console", "main.go")
	if _, err := os.Stat(consoleMain); err != nil {
		return fmt.Errorf("missing cmd/console/main.go — run: cais g console")
	}

	_, _ = fmt.Fprintln(c.Out, "=> Starting console")
	return runCmd(dir, "go", "run", "./cmd/console")
}
