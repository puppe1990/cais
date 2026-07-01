package middleware

import (
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais"
)

func ClientIP(r *http.Request, cfg cais.Config) string {
	remote := remoteAddrIP(r)
	if isTrustedProxy(remote, cfg.TrustedProxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if i := strings.Index(xff, ","); i >= 0 {
				return strings.TrimSpace(xff[:i])
			}
			return strings.TrimSpace(xff)
		}
	}
	return remote
}

func remoteAddrIP(r *http.Request) string {
	if host := r.RemoteAddr; host != "" {
		if i := strings.LastIndex(host, ":"); i >= 0 {
			return host[:i]
		}
		return host
	}
	return "127.0.0.1"
}

func isTrustedProxy(ip string, trusted []string) bool {
	for _, t := range trusted {
		if ip == t {
			return true
		}
	}
	return false
}
