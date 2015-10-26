package cnet

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type cNet struct {
	dbgName string
}

// New creates a new networking interface. The debugName is used as the
// RemoteAddr() for channel connections.
func New(debugName string) Net {
	return &cNet{
		dbgName: debugName,
	}
}

func (n *cNet) Dial(addr string, t time.Duration) (net.Conn, error) {
	prot, addr := addrSplit(addr, true)
	switch prot {
	case "ch":
		return globalChs.dial(n.dbgName, addr, t)
	default:
		return net.DialTimeout(prot, addr, t)
	}
}

func (n *cNet) HostExists(addr string) bool {
	prot, addr := addrSplit(addr, true)
	switch prot {
	case "ch":
		return true
	default:
		host, _, err := net.SplitHostPort(addr)
		if err == nil {
			addr = host
		}

		_, err = net.LookupHost(addr)

		return err == nil
	}
}

// `prot` is the protocol to use if none is specified in the addr
func (n *cNet) Resolve(prot, addr string) (net.Addr, error) {
	parts := strings.Split(addr, "://")
	if len(parts) > 1 {
		prot, addr = addrSplit(addr, false)
	}

	switch prot {
	case "ch":
		return chAddr(addr), nil
	case "tcp":
		return net.ResolveTCPAddr("tcp", addr)
	case "udp":
		return net.ResolveUDPAddr("udp", addr)
	}

	return nil, fmt.Errorf("unsupported protocol: %s", prot)
}

func (n *cNet) Listen(addr string) (net.Listener, error) {
	prot, addr := addrSplit(addr, true)
	switch prot {
	case "ch":
		return globalChs.listen(addr)
	default:
		return net.Listen(prot, addr)
	}
}

func (n *cNet) ListenPacket(addr string) (net.PacketConn, error) {
	prot, addr := addrSplit(addr, false)
	switch prot {
	case "ch":
		return globalChs.listenPacket(addr)
	default:
		return net.ListenPacket(prot, addr)
	}
}
