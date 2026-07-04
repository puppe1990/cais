package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/pwa"
)

func (c *CLI) cmdPWA(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		if isCaisFramework(mustCwd()) {
			dir = mustCwd()
		} else {
			return err
		}
	}

	for _, arg := range args {
		if arg == "--bump" {
			v, err := pwa.BumpCacheVersion(dir)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(c.Out, "→ PWA cache version bumped to %d (reinstall PWA on phone to pick up changes)\n", v)
			return nil
		}
	}

	name := filepath.Base(dir)
	if name == "." || name == "/" || name == string(os.PathSeparator) {
		name = "Cais"
	}
	_, _ = fmt.Fprintln(c.Out, "→ writing PWA assets")
	return pwa.InstallTo(dir, name)
}

func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func preferredPort(dir string) string {
	if v := resolveEnvVar(dir, "PORT"); v != "" {
		return v
	}
	return ":8080"
}

func warnPortInUse(w io.Writer, dir string) {
	port := preferredPort(dir)
	if !strings.HasPrefix(port, ":") && !strings.Contains(port, ":") {
		port = ":" + port
	}
	if cais.PortBusy(port) {
		_, _ = fmt.Fprintf(w, "=> Warning: port %s already in use — another server may be running; phone/LAN clients can hit stale code\n", port)
	}
}
