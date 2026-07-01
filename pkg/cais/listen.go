package cais

import (
	"fmt"
	"net"
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

	for i := 0; i < maxPortAttempts; i++ {
		candidate := formatListenAddr(host, base+i)
		ln, listenErr := net.Listen("tcp", candidate)
		if listenErr == nil {
			_ = ln.Close()
			return candidate, i > 0, nil
		}
	}

	return "", false, fmt.Errorf("no free port near %s after %d attempts", port, maxPortAttempts)
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

func formatListenAddr(host string, port int) string {
	if host == "" {
		return fmt.Sprintf(":%d", port)
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}
