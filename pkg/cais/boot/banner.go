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

// C: open curve · A: pointed top · I: thin column · S: classic curve
const devBannerArt = `  ____     /\      |     _____
 / ___|   /  \     |    / ____|
| |      / /\ \    |   | (___
| |___  / ____ \   |    \___ \
 \____|/_/    \_\  |_|  |____/

`
