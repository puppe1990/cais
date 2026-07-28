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
	// Detect legacy cache-first SW before Install overwrites (for messaging).
	legacySW := false
	if data, rerr := os.ReadFile(filepath.Join(dir, "web", "static", "js", "sw.js")); rerr == nil {
		legacySW = !pwa.HasNetworkFirstSPA(string(data))
	}

	_, _ = fmt.Fprintln(c.Out, "→ writing PWA assets")
	// Inertia apps: icons/manifest/SW without HTMX JS dump. HTMX apps: full InstallTo.
	// Install* uses SyncServiceWorker so CACHE_VERSION is preserved and strategy migrates (#135).
	if isInertiaApp(dir) {
		err = pwa.InstallForInertia(dir, name)
	} else {
		err = pwa.InstallTo(dir, name)
	}
	if err != nil {
		return err
	}
	// Sync again is cheap/idempotent; returns preserved CACHE_VERSION for the log line.
	_, ver, err := pwa.SyncServiceWorker(dir)
	if err != nil {
		return err
	}
	if legacySW {
		_, _ = fmt.Fprintf(c.Out, "→ sw.js updated (network-first /static/build/ + /static/css/, CACHE_VERSION=%d)\n", ver)
		_, _ = fmt.Fprintln(c.Out, "  tip: cais pwa --bump after deploy so installed PWAs drop old caches")
	} else {
		_, _ = fmt.Fprintf(c.Out, "→ sw.js network-first for SPA (CACHE_VERSION=%d)\n", ver)
	}
	return nil
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
