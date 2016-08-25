package cnet

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/iheartradio/cog/bytec"
)

type chPacketConn struct {
	mtx           sync.Mutex
	ch            *ch
	addr          chAddr
	readDeadline  time.Time
	writeDeadline time.Time
	msgs          chan chPacketMsg
	closed        bool
	closeCh       chan struct{}
}

type extChPacketConn struct {
	*chPacketConn
}

type chPacketMsg struct {
	from *chPacketConn
	msg  []byte
}

func (nc *ch) listenPacket(addr string) (net.PacketConn, error) {
	nc.mtx.Lock()
	defer nc.mtx.Unlock()

	_, ok := nc.packetConns[addr]
	if ok {
		return nil, newChError(
			false, false,
			fmt.Sprintf("udp address in use: %s", addr))
	}

	pc := &chPacketConn{
		ch:      nc,
		addr:    chAddr(addr),
		msgs:    make(chan chPacketMsg, 8),
		closeCh: make(chan struct{}),
	}

	epc := &extChPacketConn{pc}
	nc.packetConns[addr] = pc

	runtime.SetFinalizer(epc, finalizePacketConn)

	return epc, nil
}

func finalizePacketConn(epc *extChPacketConn) {
	epc.Close()
}

func (pc *extChPacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	pc.mtx.Lock()
	closed := pc.closed
	pc.mtx.Unlock()

	if closed {
		err = newChError(false, false, "socket closed")
		return
	}

	var timeout <-chan time.Time
	if !pc.readDeadline.IsZero() {
		timeout = time.After(pc.readDeadline.Sub(time.Now()))
	}

	for {
		select {
		case msg := <-pc.msgs:
			if msg.from == nil {
				continue
			}

			n = copy(b, msg.msg)
			addr = msg.from.addr
			return

		case <-timeout:
			err = newChError(true, false, "deadline passed")
			return

		case <-pc.closeCh:
			err = newChError(false, false, "socket closed")
			return
		}
	}
}

func (pc *extChPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	pc.ch.mtx.Lock()
	r, ok := pc.ch.packetConns[addr.String()]
	pc.ch.mtx.Unlock()

	if ok {
		r.mtx.Lock()

		ok = !r.closed

		r.mtx.Unlock()
	}

	if !ok {
		err = newChError(
			false, false,
			fmt.Sprintf("host not found: %s", addr.String()))
		return
	}

	pc.mtx.Lock()
	closed := pc.closed
	pc.mtx.Unlock()

	if closed {
		err = newChError(false, false, "socket closed")
		return
	}

	msg := chPacketMsg{
		from: pc.chPacketConn,
		msg:  bytec.Dup(b),
	}

	var timeout <-chan time.Time
	if !pc.writeDeadline.IsZero() {
		timeout = time.After(pc.readDeadline.Sub(time.Now()))
	}

	select {
	case r.msgs <- msg:
		n = len(b)

	case <-timeout:
		err = newChError(true, false, "deadline passed")
		return

	case <-pc.closeCh:
		err = newChError(false, false, "socket closed")

	case <-r.closeCh:
		err = newChError(
			false, false,
			fmt.Sprintf("host not found: %s", addr.String()))
	}

	return
}

func (pc *extChPacketConn) LocalAddr() net.Addr {
	return pc.addr
}

func (pc *extChPacketConn) SetDeadline(t time.Time) error {
	pc.SetReadDeadline(t)
	pc.SetWriteDeadline(t)
	return nil
}

func (pc *extChPacketConn) SetReadDeadline(t time.Time) error {
	pc.readDeadline = t
	return nil
}

func (pc *extChPacketConn) SetWriteDeadline(t time.Time) error {
	pc.writeDeadline = t
	return nil
}

func (pc *extChPacketConn) Close() error {
	runtime.SetFinalizer(pc, nil)

	func() {
		pc.ch.mtx.Lock()
		defer pc.ch.mtx.Unlock()

		delete(pc.ch.packetConns, pc.addr.String())
	}()

	func() {
		pc.mtx.Lock()
		defer pc.mtx.Unlock()

		if !pc.closed {
			pc.closed = true
			close(pc.closeCh)
		}
	}()

	return nil
}
