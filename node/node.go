// Package node provides information about the local node
package node

import (
	"fmt"
	"net"
)

// ResolveListenAddress validates the given address and resolves any special
// hostnames to actual addresses.
//
// Currently, it resolves the following:
//     - <ec2-priv>: resolves to the node's private IP address
func ResolveListenAddress(addr *string) error {
	return resolveListenAddress(addr, &EC2)
}

func resolveListenAddress(addr *string, ec2 *EC2Metadata) error {
	if len(*addr) == 0 {
		return fmt.Errorf("no listen address given")
	}

	host, port, err := net.SplitHostPort(*addr)
	if err != nil {
		return fmt.Errorf("invalid listen address: %s: %v", *addr, err)
	}

	switch host {
	case "<ec2-priv>":
		host, err = ec2.GetPrivIP()
	}

	if err != nil {
		err = fmt.Errorf("while resolving hostname %s: %v", *addr, err)
	} else {
		*addr = fmt.Sprintf("%s:%s", host, port)
	}

	return err
}
