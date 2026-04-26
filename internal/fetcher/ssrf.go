package fetcher

import (
	"net"
	"strings"
)

func isBlockedHostname(host string) bool {
	value := strings.ToLower(strings.TrimSpace(host))
	return value == "localhost" || strings.HasSuffix(value, ".local")
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}
