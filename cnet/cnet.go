// Package cnet provides more network utils, including sockets made from
// channels.
//
// Addresses for this package are standardized as follows:
//
//     tcp:// - operate in TCP space
//     ucp:// - operate in UDP space
//      ch:// - operate with named, process-local, channel-based sockets
package cnet

import (
	"net"
	"time"
)

// Net wraps golang's standard net package, providing process-local channel
// connections.
type Net interface {
	// Analogue of net.Dial. Addresses should be of the form "tcp://address",
	// "udp://address", or "ch://channel-name", depending on the type of socket
	// you want to create.
	Dial(addr string, timeout time.Duration) (net.Conn, error)

	// Check if a hostname exists on the network
	HostExists(addr string) bool

	// Resolve the given address. If no protocol is given in the address (in the
	// form "tcp://address"), then protocol is used to determine which protocol
	// to resolve for.
	Resolve(protocol, addr string) (net.Addr, error)

	// Analogue of net.Listen. Address rules are the same as for Dial().
	Listen(addr string) (net.Listener, error)

	// Analogue of net.ListenPacket. Address rules are the same as for Dial().
	ListenPacket(addr string) (net.PacketConn, error)
}
