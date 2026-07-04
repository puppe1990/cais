package boot

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

func PrintDevBanner(w io.Writer, version string) {
	version = strings.TrimSpace(version)
	if version == "" {
		version = "dev"
	}
	_, _ = fmt.Fprint(w, devBannerArt)
	_, _ = fmt.Fprintf(w, "Cais v%s · %s · hot reload: Go (air) + HTML (~400ms) + static (disk) + Tailwind\n\n", version, runtime.Version())
}

const devSeedWarning = "=> Demo user may exist: demo@example.com / password (development only)"

func WriteDevSeedWarning(w io.Writer, env string) {
	if env == "development" {
		_, _ = fmt.Fprintln(w, devSeedWarning)
	}
}

const devBannerArt = ` ██████╗ █████╗ ██╗███████╗
██╔════╝██╔══██╗██║██╔════╝
██║     ███████║██║███████╗
██║     ██╔══██║██║╚════██║
╚██████╗██║  ██║██║███████║
 ╚═════╝╚═╝  ╚═╝╚═╝╚══════╝

`
