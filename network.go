package xmpp

import (
	"strconv"
	"strings"
)

// ensurePort adds a port to an address if none are provided.
// It handles both IPV4 and IPV6 addresses.
func ensurePort(addr string, port int) string {
	// This is an IPV6 address literal
	if strings.HasPrefix(addr, "[") {
		// if address has no port (behind his ipv6 address) - add default port
		if strings.LastIndex(addr, ":") <= strings.LastIndex(addr, "]") {
			return addr + ":" + strconv.Itoa(port)
		}
		return addr
	}

	// This is either an IPV6 address without bracket or an IPV4 address
	switch strings.Count(addr, ":") {
	case 0:
		// This is IPV4 without port
		return addr + ":" + strconv.Itoa(port)
	case 1:
		// This is IPV$ with port
		return addr
	default:
		// This is IPV6 without port, as you need to use bracket with port in IPV6
		return "[" + addr + "]:" + strconv.Itoa(port)
	}
}
