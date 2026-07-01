package cli

import (
	"fmt"
	"io"
	"os"
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
	case "install", "i":
		return c.cmdInstall()
	case "css":
		return c.cmdCSS()
	case "dev":
		return c.cmdDev()
	case "build", "b":
		return c.cmdBuild()
	case "server", "s":
		return c.cmdServer()
	case "test":
		return c.cmdTest()
	case "doctor":
		return c.cmdDoctor()
	case "console", "c":
		return c.cmdConsole()
	case "db":
		return c.cmdDB(args[1:])
	case "routes":
		return c.cmdRoutes()
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
  cais new <app> [dir] [--minimal] [--blank] [--module <path>]
                               Create a new app (default dir: ./<app>)
  cais new <app> [dir] --minimal   Slim app (home only)
  cais new <app> [dir] --blank     Empty app (no starter content)
  cais new <app> [dir] --module <path>   Override go module path
  cais g [--dry-run] handler <name>      Generate handler + test + page template
  cais g [--dry-run] resource <name> [--fields title:string,url:url] [--public] [--paginate] [--no-seed] [--admin-auth session|bearer]
  cais g [--dry-run] model <name> [--fields title:string,url:url]
  cais g [--dry-run] page <name>         Generate page template only
  cais g [--dry-run] migration <name>    Generate SQL migration file
  cais g [--dry-run] auth                Add login/logout and protect dashboard
  cais g ci                  Add GitHub Actions CI, pre-commit, lint, Prettier
  cais install               npm install + go mod tidy
  cais css                   Build Tailwind CSS
  cais dev                   Hot reload (air + tailwind watch)
  cais build                 Build bin/server
  cais server                Run the app (go run ./cmd/server)
  cais test                  Run tests (go test ./...)
  cais doctor                Check app setup (htmx, air, go.mod)
  cais console               Interactive app console (Go REPL + SQL)
  cais db migrate            Run pending SQL migrations
  cais db status             List migration status
  cais db rollback           Remove last applied migration record (no SQL down)
  cais db prune-sessions     Delete expired login sessions from SQLite
  cais routes                List HTTP routes from internal/app/routes.go
  cais help                  Show this help

Aliases:
  cais g        → cais generate
  cais i        → cais install
  cais b        → cais build
  cais s        → cais server
  cais c        → cais console

Examples:
  cais new myapp && cd myapp && cais install && cais dev
  cais g handler settings
  cais console
  cais css && cais server`)
}

func (c *CLI) cmdNew(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cais new <app> [dir] [--minimal] [--blank] [--module <path>]")
	}

	opts, err := parseNewArgs(args)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(opts.dir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(abs); err == nil {
		return fmt.Errorf("directory %s already exists", abs)
	}

	module := opts.module
	if module == "" {
		module = moduleName(opts.name)
	}
	if err := scaffoldNewApp(abs, scaffoldData{
		AppName:    opts.name,
		ModulePath: module,
	}, opts.minimal, opts.blank); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(c.Out, "Created app %q at %s\n\nNext steps:\n  cd %s\n  cais install\n  cais dev\n", opts.name, abs, abs)
	return nil
}

type newOpts struct {
	name    string
	dir     string
	minimal bool
	blank   bool
	module  string
}

func parseNewArgs(args []string) (newOpts, error) {
	opts := newOpts{}
	positional := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--minimal":
			opts.minimal = true
		case "--blank":
			opts.blank = true
		case "--module":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--module requires a value")
			}
			i++
			opts.module = args[i]
		default:
			positional = append(positional, args[i])
		}
	}
	if len(positional) == 0 {
		return opts, fmt.Errorf("usage: cais new <app> [dir] [--minimal] [--blank] [--module <path>]")
	}

	opts.name = positional[0]
	opts.dir = opts.name
	if len(positional) > 1 {
		opts.dir = positional[1]
	}
	return opts, nil
}

func (c *CLI) cmdGenerate(args []string) error {
	dryRun := false
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			continue
		}
		filtered = append(filtered, arg)
	}
	args = filtered

	if len(args) < 1 {
		return fmt.Errorf("usage: cais g [--dry-run] <handler|page|migration|resource|model|console|auth|ci> [name]")
	}

	kind := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !isCaisApp(cwd) {
		if isCaisFramework(cwd) {
			return fmt.Errorf("you are inside the Cais framework directory — cd into your app first")
		}
		return fmt.Errorf("not a Cais app (missing go.mod with github.com/puppe1990/cais as a dependency)")
	}

	switch kind {
	case "console":
		return scaffoldConsole(cwd)
	case "auth":
		return scaffoldAuth(cwd, scaffoldData{AppName: filepath.Base(cwd), ModulePath: moduleFromDir(cwd)}, dryRun)
	case "ci":
		return scaffoldCI(cwd, scaffoldData{AppName: filepath.Base(cwd), ModulePath: moduleFromDir(cwd)})
	case "handler", "page", "migration", "resource", "model":
		if len(args) < 2 {
			return fmt.Errorf("usage: cais g %s <name>", kind)
		}
		name := args[1]
		switch kind {
		case "handler":
			return scaffoldHandler(cwd, name, dryRun)
		case "page":
			return scaffoldPage(cwd, name, dryRun)
		case "migration":
			return scaffoldMigration(cwd, name, dryRun)
		case "resource":
			opts, err := parseResourceOpts(args[2:])
			if err != nil {
				return err
			}
			opts.dryRun = dryRun
			return scaffoldResource(cwd, name, opts)
		case "model":
			opts, err := parseModelOpts(args[2:])
			if err != nil {
				return err
			}
			opts.dryRun = dryRun
			return scaffoldModel(cwd, name, opts)
		}
	default:
		return fmt.Errorf("unknown generator %q (use handler, page, migration, resource, model, auth, ci, or console)", kind)
	}
	return nil
}

func (c *CLI) cmdServer() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(c.Out, "→ go run ./cmd/server")
	return runCmd(dir, "go", "run", "./cmd/server")
}

func (c *CLI) cmdDoctor() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	return runDoctor(c.Out, dir)
}

func (c *CLI) cmdTest() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	return runCmd(dir, "go", "test", "./...", "-race", "-count=1")
}

func moduleName(app string) string {
	slug := strings.ToLower(strings.ReplaceAll(app, "-", ""))
	return "github.com/puppe1990/" + slug
}

func moduleFromDir(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return moduleName(filepath.Base(dir))
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return moduleName(filepath.Base(dir))
}

func isCaisApp(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	content := string(data)
	if strings.HasPrefix(content, "module github.com/puppe1990/cais") {
		return false
	}
	return strings.Contains(content, "github.com/puppe1990/cais")
}

func isCaisFramework(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	return strings.HasPrefix(string(data), "module github.com/puppe1990/cais")
}
