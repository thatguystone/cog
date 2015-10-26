package cnet

import "strings"

func addrSplit(addr string, tcp bool) (string, string) {
	parts := strings.Split(addr, "://")

	if len(parts) == 1 {
		if tcp {
			return "tcp", addr
		}

		return "udp", addr
	}

	return parts[0], parts[1]
}
