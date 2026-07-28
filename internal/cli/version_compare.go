package cli

import (
	"strconv"
	"strings"
)

// semverCore is major.minor.patch parsed from a version string (v-prefix and
// build/pre-release suffixes ignored). Used to compare CLI vs go.mod.
type semverCore struct {
	Major, Minor, Patch int
	OK                  bool
}

func parseSemverCore(s string) semverCore {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if s == "" || s == "dev" || s == "(devel)" {
		return semverCore{}
	}
	// Drop pseudo-version / pre-release: 0.6.1-0.2026... or 0.8.0-rc.1
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return semverCore{}
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	pat := 0
	var err3 error
	if len(parts) >= 3 {
		pat, err3 = strconv.Atoi(parts[2])
	}
	if err1 != nil || err2 != nil || err3 != nil {
		return semverCore{}
	}
	return semverCore{Major: maj, Minor: min, Patch: pat, OK: true}
}

// compareSemverCore returns -1 if a < b, 0 if equal, 1 if a > b.
// Non-OK values compare as equal (caller should skip).
func compareSemverCore(a, b semverCore) int {
	if !a.OK || !b.OK {
		return 0
	}
	if a.Major != b.Major {
		if a.Major < b.Major {
			return -1
		}
		return 1
	}
	if a.Minor != b.Minor {
		if a.Minor < b.Minor {
			return -1
		}
		return 1
	}
	if a.Patch != b.Patch {
		if a.Patch < b.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// minViteWatchVersion is the first release with cais dev → vite build --watch (#128).
const minViteWatchVersion = "0.8.0"
