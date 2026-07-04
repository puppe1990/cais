package pwa

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var cacheVersionRe = regexp.MustCompile(`const CACHE_VERSION = (\d+);`)

// BumpCacheVersion increments CACHE_VERSION in web/static/js/sw.js so installed PWAs fetch fresh assets.
func BumpCacheVersion(appDir string) (int, error) {
	path := filepath.Join(appDir, "web", "static", "js", "sw.js")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read sw.js: %w", err)
	}
	m := cacheVersionRe.FindStringSubmatch(string(data))
	if m == nil {
		return 0, fmt.Errorf("sw.js missing const CACHE_VERSION — run cais pwa or make pwa")
	}
	cur, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, err
	}
	next := cur + 1
	updated := cacheVersionRe.ReplaceAllString(string(data), fmt.Sprintf("const CACHE_VERSION = %d;", next))
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return 0, err
	}
	return next, nil
}
