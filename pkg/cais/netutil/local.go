package netutil

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// LANURLs returns http URLs for non-loopback IPv4 interfaces on the listen port.
func LANURLs(port string) []string {
	p, err := parsePortNumber(port)
	if err != nil {
		return nil
	}

	var urls []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil || ipNet.IP.IsLoopback() {
				continue
			}
			urls = append(urls, fmt.Sprintf("http://%s:%d", ipNet.IP.String(), p))
		}
	}
	return urls
}

func parsePortNumber(port string) (int, error) {
	port = strings.TrimSpace(port)
	if port == "" {
		return 8080, nil
	}
	if strings.HasPrefix(port, ":") {
		return strconv.Atoi(port[1:])
	}
	_, portPart, err := net.SplitHostPort(port)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(portPart)
}
