package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CLI struct {
	Out io.Writer
}

func Main() int {
	c := &CLI{Out: os.Stdout}
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "cais: %v\n", err)
		return 1
	}
	return 0
}

func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		c.printHelp()
		return nil
	}

	switch args[0] {
	case "new":
		return c.cmdNew(args[1:])
	case "generate", "g":
		return c.cmdGenerate(args[1:])
	case "server", "s":
		return c.cmdServer()
	case "test":
		return c.cmdTest()
	case "help", "-h", "--help":
		c.printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q (run cais help)", args[0])
	}
}

func (c *CLI) printHelp() {
	_, _ = fmt.Fprintln(c.Out, `Cais — Rails-style CLI for Go full-stack apps

Usage:
  cais new <app> [dir]       Create a new app (default dir: ./<app>)
  cais g handler <name>      Generate handler + test + page template
  cais g page <name>         Generate page template only
  cais g migration <name>    Generate SQL migration file
  cais server                Run the app (go run ./cmd/server)
  cais test                  Run tests (go test ./...)
  cais help                  Show this help

Aliases:
  cais g        → cais generate
  cais s        → cais server

Examples:
  cais new dashboard ../dashboard
  cais g handler settings
  cais server`)
}

func (c *CLI) cmdNew(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cais new <app> [dir]")
	}

	name := args[0]
	dir := name
	if len(args) > 1 {
		dir = args[1]
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(abs); err == nil {
		return fmt.Errorf("directory %s already exists", abs)
	}

	module := moduleName(name)
	if err := scaffoldNewApp(abs, scaffoldData{
		AppName:    name,
		ModulePath: module,
	}); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(c.Out, "Created app %q at %s\n\nNext steps:\n  cd %s\n  npm install\n  make dev\n", name, abs, abs)
	return nil
}

func (c *CLI) cmdGenerate(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: cais g <handler|page|migration> <name>")
	}

	kind := args[0]
	name := args[1]
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !isCaisApp(cwd) {
		return fmt.Errorf("not a Cais app (missing go.mod with github.com/matheuspuppe/cais)")
	}

	switch kind {
	case "handler":
		return scaffoldHandler(cwd, name)
	case "page":
		return scaffoldPage(cwd, name)
	case "migration":
		return scaffoldMigration(cwd, name)
	default:
		return fmt.Errorf("unknown generator %q (use handler, page, or migration)", kind)
	}
}

func (c *CLI) cmdServer() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if !isCaisApp(cwd) {
		return fmt.Errorf("not a Cais app")
	}
	cmd := exec.Command("go", "run", "./cmd/server")
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c *CLI) cmdTest() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "test", "./...", "-race", "-count=1")
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func moduleName(app string) string {
	slug := strings.ToLower(strings.ReplaceAll(app, "-", ""))
	return "github.com/puppe1990/" + slug
}

func isCaisApp(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "github.com/matheuspuppe/cais")
}
