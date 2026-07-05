package cli

import (
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais/pwa"
)

// maybeBumpDevCache increments CACHE_VERSION when sw.js exists (phone testing during cais dev).
func maybeBumpDevCache(dir string) (bumped bool, version int, err error) {
	path := filepath.Join(dir, "web", "static", "js", "sw.js")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}
	v, err := pwa.BumpCacheVersion(dir)
	if err != nil {
		return false, 0, err
	}
	return true, v, nil
}
