package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais/boot"
)

const (
	cssInput  = "input.css"
	cssOutput = "web/static/css/styles.css"
	serverBin = "bin/server"
)

func frameworkVersion() string {
	return boot.CaisVersion()
}

func (c *CLI) appDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if !isCaisApp(cwd) {
		if isCaisFramework(cwd) {
			return "", fmt.Errorf("you are inside the Cais framework directory — cd into your app first")
		}
		return "", fmt.Errorf("not a Cais app (run from app root with go.mod)")
	}
	return cwd, nil
}

func (c *CLI) cmdInstall() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		_, _ = fmt.Fprintln(c.Out, "→ npm install")
		if err := runCmd(dir, "npm", "install"); err != nil {
			return fmt.Errorf("npm install: %w", err)
		}
	}

	_, _ = fmt.Fprintln(c.Out, "→ go mod tidy")
	if err := runCmd(dir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	_, _ = fmt.Fprintln(c.Out, "Done. Run: cais css && cais dev")
	return nil
}

func (c *CLI) cmdCSS() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(c.Out, "→ tailwind build")
	return runTailwindBuild(dir, false)
}

func (c *CLI) cmdBuild() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	if err := runTailwindBuild(dir, false); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(c.Out, "→ go build")
	return runCmd(dir, "go", "build", "-ldflags=-s -w", "-o", serverBin, "./cmd/server")
}

func (c *CLI) cmdDev() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}

	if err := runTailwindBuild(dir, false); err != nil {
		return err
	}

	watch := exec.Command("npx", "tailwindcss", "-i", cssInput, "-o", cssOutput, "--watch")
	watch.Dir = dir
	watch.Stdout = os.Stdout
	watch.Stderr = os.Stderr
	if err := watch.Start(); err != nil {
		return fmt.Errorf("tailwind watch: %w", err)
	}
	defer func() { _ = watch.Process.Kill() }()

	if air := findAir(); air != "" {
		boot.PrintDevBanner(c.Out, boot.CaisVersion())
		return runCmd(dir, air, "-c", ".air.toml")
	}

	_, _ = fmt.Fprintln(c.Out, "=> Starting dev server (go run; install air for hot reload)")
	return runCmd(dir, "go", "run", "./cmd/server")
}

func runTailwindBuild(dir string, watch bool) error {
	if _, err := os.Stat(filepath.Join(dir, cssInput)); err != nil {
		return fmt.Errorf("missing %s", cssInput)
	}
	args := []string{"tailwindcss", "-i", cssInput, "-o", cssOutput}
	if watch {
		args = append(args, "--watch")
	} else {
		args = append(args, "--minify")
	}
	return runCmd(dir, "npx", args...)
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func findAir() string {
	if path, err := exec.LookPath("air"); err == nil {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	candidate := filepath.Join(home, "go", "bin", "air")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}
