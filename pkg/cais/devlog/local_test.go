package devlog

import (
	"net/http/httptest"
	"testing"
)

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       bool
	}{
		{name: "ipv4 loopback", remoteAddr: "127.0.0.1:1234", want: true},
		{name: "ipv6 loopback", remoteAddr: "[::1]:1234", want: true},
		{name: "public ipv4", remoteAddr: "203.0.113.1:1234", want: false},
		{name: "public ipv6", remoteAddr: "[2001:db8::1]:1234", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if got := IsLoopback(req); got != tt.want {
				t.Fatalf("IsLoopback(%q) = %v, want %v", tt.remoteAddr, got, tt.want)
			}
		})
	}
}
