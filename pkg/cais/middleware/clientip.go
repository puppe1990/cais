package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
)

// ClientIP returns the client address. X-Forwarded-For is trusted only when RemoteAddr is in TRUSTED_PROXIES.
// Without the allowlist, clients can spoof XFF and evade rate limits.
//
// XFF is walked right-to-left: append-style proxies leave client-supplied
// prefixes intact, so the leftmost entry is spoofable (#170). Trusted hops are
// skipped; the first non-trusted address is returned.
func ClientIP(r *http.Request, cfg cais.Config) string {
	remote := remoteAddrIP(r)
	if !isTrustedProxy(remote, cfg.TrustedProxies) {
		return remote
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(parts[i])
			if candidate != "" && !isTrustedProxy(candidate, cfg.TrustedProxies) {
				return candidate
			}
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	return remote
}

func remoteAddrIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		if r.RemoteAddr != "" {
			return strings.TrimPrefix(strings.TrimSuffix(r.RemoteAddr, "]"), "[")
		}
		return "127.0.0.1"
	}
	return host
}

func isTrustedProxy(ip string, trusted []string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, t := range trusted {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if strings.Contains(t, "/") {
			_, network, err := net.ParseCIDR(t)
			if err != nil {
				continue
			}
			if network.Contains(parsed) {
				return true
			}
			continue
		}
		if t == ip {
			return true
		}
	}
	return false
}
