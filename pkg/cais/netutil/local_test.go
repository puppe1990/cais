package netutil

import (
	"strings"
	"testing"
)

func TestLANURLs_formatsPort(t *testing.T) {
	urls := LANURLs(":8080")
	for _, u := range urls {
		if !strings.HasSuffix(u, ":8080") {
			t.Errorf("LANURL %q should end with :8080", u)
		}
		if !strings.HasPrefix(u, "http://") {
			t.Errorf("LANURL %q should start with http://", u)
		}
	}
}

func TestLANURLs_skipsLoopback(t *testing.T) {
	for _, u := range LANURLs(":8080") {
		if strings.Contains(u, "127.0.0.1") || strings.Contains(u, "localhost") {
			t.Errorf("LANURLs should not include loopback: %q", u)
		}
	}
}
