package cais

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const maxPortAttempts = 20

// ResolvePort returns a listen address, shifting to the next free port in development
// when the preferred address is already in use.
func ResolvePort(port, env string) (resolved string, shifted bool, err error) {
	port = strings.TrimSpace(port)
	if port == "" {
		port = ":8080"
	}
	if env != "development" {
		return port, false, nil
	}

	host, base, err := parseListenPort(port)
	if err != nil {
		return "", false, err
	}

	attempts := maxPortAttempts
	if portStrictEnabled() {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		candidate := formatListenAddr(host, base+i)
		ln, listenErr := net.Listen("tcp", candidate)
		if listenErr == nil {
			_ = ln.Close()
			return candidate, i > 0, nil
		}
		lastErr = listenErr
	}

	if portStrictEnabled() && lastErr != nil {
		return "", false, fmt.Errorf("port %s in use (%v); stop the other process or run your app stop script", port, lastErr)
	}

	return "", false, fmt.Errorf("no free port near %s after %d attempts", port, attempts)
}

func portStrictEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("PORT_STRICT")))
	return v == "1" || v == "true" || v == "yes"
}

func parseListenPort(port string) (host string, base int, err error) {
	if strings.HasPrefix(port, ":") {
		base, err = strconv.Atoi(port[1:])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port %q", port)
		}
		return "", base, nil
	}

	hostPart, portPart, err := net.SplitHostPort(port)
	if err != nil {
		return "", 0, fmt.Errorf("invalid listen address %q", port)
	}
	base, err = strconv.Atoi(portPart)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in %q", port)
	}
	return hostPart, base, nil
}

// PortBusy reports whether something is already listening on addr.
func PortBusy(addr string) bool {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

func formatListenAddr(host string, port int) string {
	if host == "" {
		return fmt.Sprintf(":%d", port)
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}
