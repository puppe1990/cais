package pwa

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasNetworkFirstSPA(t *testing.T) {
	legacy := `if (url.pathname.startsWith("/static/")) {
  caches.match(request)
}`
	if HasNetworkFirstSPA(legacy) {
		t.Error("legacy cache-first should not report network-first SPA")
	}
	modern, err := assets.ReadFile("assets/sw.js")
	if err != nil {
		t.Fatal(err)
	}
	if !HasNetworkFirstSPA(string(modern)) {
		t.Error("embedded sw.js should be network-first for /static/build/")
	}
}

func TestSyncServiceWorker_migratesLegacy(t *testing.T) {
	dir := t.TempDir()
	jsDir := filepath.Join(dir, "web", "static", "js")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `const CACHE_VERSION = 4;
const CACHE = "cais-static-v" + CACHE_VERSION;
self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);
  if (url.pathname.startsWith("/static/")) {
    event.respondWith(caches.match(event.request));
  }
});
`
	if err := os.WriteFile(filepath.Join(jsDir, "sw.js"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	updated, ver, err := SyncServiceWorker(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Error("expected updated=true when migrating legacy SW")
	}
	if ver != 4 {
		t.Errorf("version = %d, want preserved 4", ver)
	}
	body, err := os.ReadFile(filepath.Join(jsDir, "sw.js"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !HasNetworkFirstSPA(s) {
		t.Error("migrated sw.js missing network-first SPA")
	}
	if !strings.Contains(s, "CACHE_VERSION = 4") {
		t.Errorf("should preserve CACHE_VERSION=4, got:\n%s", s)
	}
}

func TestSyncServiceWorker_idempotentStrategy(t *testing.T) {
	dir := t.TempDir()
	if err := WriteStaticInertia(dir, DefaultConfig("app")); err != nil {
		t.Fatal(err)
	}
	// Second sync: still ok, may rewrite same content
	_, ver, err := SyncServiceWorker(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ver < 1 {
		t.Errorf("version = %d", ver)
	}
	body, _ := os.ReadFile(filepath.Join(dir, "web/static/js/sw.js"))
	if !HasNetworkFirstSPA(string(body)) {
		t.Error("sw.js should keep network-first after re-sync")
	}
}
