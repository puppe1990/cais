package devlog

import (
	"net"
	"net/http"
)

func LocalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsLoopback(r) {
			http.Error(w, "logs only available on localhost", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func IsLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
