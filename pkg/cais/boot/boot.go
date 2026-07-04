package boot

import (
	"fmt"
	"io"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/netutil"
)

type Options struct {
	AppName         string
	Config          cais.Config
	Version         string
	PortShiftedFrom string
}

func Print(w io.Writer, opts Options) {
	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version = "dev"
	}
	app := strings.TrimSpace(opts.AppName)
	if app == "" {
		app = "app"
	}

	if from := strings.TrimSpace(opts.PortShiftedFrom); from != "" && from != opts.Config.Port {
		_, _ = fmt.Fprintf(w, "=> Port %s in use, using %s\n", from, opts.Config.Port)
	}
	_, _ = fmt.Fprintf(w, "=> Booting %s (Cais v%s)\n", app, version)
	_, _ = fmt.Fprintf(w, "=> Environment: %s\n", opts.Config.Env)
	_, _ = fmt.Fprintf(w, "=> Database:    sqlite3 (%s)\n", opts.Config.DBPath)
	_, _ = fmt.Fprintf(w, "=> Listening on %s\n", ListenURL(opts.Config.Port))
	for _, url := range netutil.LANURLs(opts.Config.Port) {
		_, _ = fmt.Fprintf(w, "=> LAN:          %s\n", url)
	}
	WriteDevSeedWarning(w, opts.Config.Env)
	_, _ = fmt.Fprintln(w, "=> Ctrl-C to stop")
}

func ListenURL(port string) string {
	port = strings.TrimSpace(port)
	if port == "" {
		port = ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return "http://127.0.0.1" + port
	}
	if strings.Contains(port, "://") {
		return port
	}
	return "http://" + port
}
