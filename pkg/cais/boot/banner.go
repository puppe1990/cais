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
	_, _ = fmt.Fprintf(w, "Cais v%s · %s · hot reload\n\n", version, runtime.Version())
}

// Four distinct letters: C · A · I · S
const devBannerArt = `  ____    ___   _____   ____
 / ___|  / _ \ |_   _| / ___|
| |     | | | |  | |   \___ \
| |___  | |_| |  | |    ___) |
 \____|  \___/   |_|   |____/

`
