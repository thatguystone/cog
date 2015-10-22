package cnet

import "net"

// RemoveDefaultPort strips the port from an address, assuming the port matches
// defaultPort. This is most useful for turning an address like "example.com:80"
// into "example.com", where the default port for HTTP (80) is uncessary.
func RemoveDefaultPort(addr string, defaultPort string) string {
	host, port, _ := net.SplitHostPort(addr)
	if port == defaultPort {
		addr = host
	}

	return addr
}
