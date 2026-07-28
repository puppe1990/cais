package pwa

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// HasNetworkFirstSPA reports whether sw.js uses network-first for SPA build
// assets (/static/build/). False for legacy cache-first-all-/static/ workers.
func HasNetworkFirstSPA(swBody string) bool {
	return strings.Contains(swBody, "/static/build/") &&
		(strings.Contains(swBody, "networkFirst") || strings.Contains(swBody, "network-first"))
}

// SyncServiceWorker overwrites web/static/js/sw.js with the embedded current
// template (network-first for /static/build/ and /static/css/). Preserves the
// existing CACHE_VERSION when present so a follow-up bump is optional.
// Returns true when the file was written or strategy upgraded.
func SyncServiceWorker(appDir string) (updated bool, version int, err error) {
	path := filepath.Join(appDir, "web", "static", "js", "sw.js")
	prevVersion := 1
	prevBody := ""
	if data, readErr := os.ReadFile(path); readErr == nil {
		prevBody = string(data)
		if m := cacheVersionRe.FindStringSubmatch(prevBody); m != nil {
			if v, aerr := strconv.Atoi(m[1]); aerr == nil {
				prevVersion = v
			}
		}
		// Already network-first: still rewrite so apps pick up other SW fixes; version kept.
	}

	tpl, err := assets.ReadFile("assets/sw.js")
	if err != nil {
		return false, 0, fmt.Errorf("embed sw.js: %w", err)
	}
	body := string(tpl)
	if prevVersion > 1 {
		body = cacheVersionRe.ReplaceAllString(body, fmt.Sprintf("const CACHE_VERSION = %d;", prevVersion))
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, 0, err
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return false, 0, err
	}

	// updated when missing, strategy changed, or content differed after version normalize
	updated = prevBody == "" || !HasNetworkFirstSPA(prevBody) || normalizeSW(prevBody) != normalizeSW(body)
	return updated, prevVersion, nil
}

func normalizeSW(s string) string {
	// Compare strategy/content ignoring CACHE_VERSION line for "already correct" checks.
	return cacheVersionRe.ReplaceAllString(s, "const CACHE_VERSION = N;")
}
