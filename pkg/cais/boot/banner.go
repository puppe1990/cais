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

const devBannerArt = `   ____      _
  / ___|__ _(_) ___
 | |   / _` + "`" + `_ | |/ __|
 | |__| (_| | | (__
  \____\__,_|_|\___|

`
