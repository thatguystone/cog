package cnet

import (
	"errors"
	"net"
	"runtime"
	"sync"
	"time"
)

// XNet provides extensions on top of basic Net. This may be used in place of a
// Net for testing.
type XNet interface {
	Net

	// SetOffline sets this net interface as offline. Any incoming connections
	// are silently closed. Any existing connections are closed. Any attempts to
	// dial out are met with an error. Any data sent/received over a PacketConn
	// is silently dropped.
	SetOffline(bool)
}

type xNet struct {
	cNet

	rwmtx   sync.RWMutex
	offline bool
	conns   map[net.Conn]struct{}
}

type xConn struct {
	net.Conn
	x *xNet
}

type xListener struct {
	net.Listener
	x *xNet
}

type xPacketConn struct {
	net.PacketConn
	x *xNet
}

// Returned from some operations when SetOffline(true) is called.
var ErrOffline = errors.New("net is currently offline")

// NewX creates a new instance of XNet.
func NewX(debugName string) XNet {
	return &xNet{
		cNet: cNet{
			dbgName: debugName,
		},
		conns: map[net.Conn]struct{}{},
	}
}

func (x *xNet) isOffline() bool {
	x.rwmtx.RLock()
	defer x.rwmtx.RUnlock()

	return x.offline
}

func (x *xNet) SetOffline(offline bool) {
	x.rwmtx.Lock()
	defer x.rwmtx.Unlock()

	x.offline = offline

	if offline {
		conns := x.conns
		x.conns = map[net.Conn]struct{}{}

		for c := range conns {
			c.Close()
		}
	}
}

func (x *xNet) Dial(addr string, t time.Duration) (c net.Conn, err error) {
	if x.isOffline() {
		err = ErrOffline
	}

	if err == nil {
		c, err = x.cNet.Dial(addr, t)
	}

	x.rwmtx.Lock()
	defer x.rwmtx.Unlock()

	if x.offline {
		err = ErrOffline
		c = nil
	}

	if err == nil {
		x.conns[c] = struct{}{}

		c = &xConn{
			Conn: c,
			x:    x,
		}
		runtime.SetFinalizer(c, finalizeXConn)
	}

	return
}

func (x *xNet) HostExists(addr string) bool {
	if x.isOffline() {
		return false
	}

	return x.cNet.HostExists(addr)
}

func (x *xNet) Resolve(prot, addr string) (net.Addr, error) {
	if x.isOffline() {
		return nil, ErrOffline
	}

	return x.cNet.Resolve(prot, addr)
}

func (x *xNet) Listen(addr string) (net.Listener, error) {
	l, err := x.cNet.Listen(addr)

	if err == nil {
		l = &xListener{
			Listener: l,
			x:        x,
		}
	}

	return l, err
}

func (x *xNet) ListenPacket(addr string) (net.PacketConn, error) {
	c, err := x.cNet.ListenPacket(addr)

	if err == nil {
		c = xPacketConn{
			PacketConn: c,
			x:          x,
		}
	}

	return c, err
}

func finalizeXConn(c *xConn) {
	c.x.rwmtx.Lock()
	delete(c.x.conns, c.Conn)
	c.x.rwmtx.Unlock()

	c.Close()
}

func (l *xListener) Accept() (c net.Conn, err error) {
	first := true

	for first || (l.x.isOffline() && err == nil) {
		if c != nil {
			c.Close()
		}

		first = false
		c, err = l.Listener.Accept()
	}

	return
}

func (pc xPacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	first := true

	for first || (pc.x.isOffline() && err == nil) {
		first = false
		n, addr, err = pc.PacketConn.ReadFrom(b)
	}

	return
}

func (pc xPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if !pc.x.isOffline() {
		n, err = pc.PacketConn.WriteTo(b, addr)
	} else {
		// Packet connections don't know remote state, so it should look like
		// send succeeds
		n = len(b)
	}

	return
}
