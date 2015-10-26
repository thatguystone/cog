package cnet

import (
	"net"
	"runtime"
	"testing"
	"time"
)

func TestChPacket(t *testing.T) {
	addr, c, nc := newChTest(t)

	l, err := nc.listenPacket(addr)
	c.MustNotError(err, "failed to listen")
	defer l.Close()
}

func TestChPacketGC(t *testing.T) {
	addr, c, nc := newChTest(t)

	l, err := nc.listenPacket(addr)
	c.MustNotError(err, "failed to listen")

	ll := l.(*extChPacketConn).chPacketConn
	c.Until(time.Second, func() bool {
		runtime.GC()
		return ll.closed
	})
}

func TestChPacketAddressInUse(t *testing.T) {
	addr, c, nc := newChTest(t)

	l, err := nc.listenPacket(addr)
	c.MustNotError(err, "failed to listen")
	defer l.Close()

	_, err = nc.listenPacket(addr)
	c.MustError(err, "should fail to listen when address is in use")
}

func TestChPacketReadWrite(t *testing.T) {
	addr, c, nc := newChTest(t)

	l0, err := nc.listenPacket(addr + "0")
	c.MustNotError(err, "failed to listen")
	defer l0.Close()

	l1, err := nc.listenPacket(addr + "1")
	c.MustNotError(err, "failed to listen")
	defer l1.Close()

	l0.WriteTo([]byte("test"), l1.LocalAddr())

	b := [4]byte{}
	_, raddr, err := l1.ReadFrom(b[:])
	c.MustNotError(err, "failed to read from remote")
	c.Equal(l0.LocalAddr(), raddr)
	c.Equal("test", string(b[:]))
}

func TestChPacketReadWriteClosed(t *testing.T) {
	addr, c, nc := newChTest(t)

	l0, err := nc.listenPacket(addr + "0")
	c.MustNotError(err, "failed to listen")
	defer l0.Close()

	l1, err := nc.listenPacket(addr + "1")
	c.MustNotError(err, "failed to listen")
	l1.Close()

	_, err = l0.WriteTo([]byte("test"), l1.LocalAddr())
	c.MustError(err, "should be disconnected")

	_, err = l1.WriteTo([]byte("test"), l0.LocalAddr())
	c.MustError(err, "should be disconnected")

	_, _, err = l1.ReadFrom([]byte("test"))
	c.MustError(err, "should be disconnected")

	msgs := l0.(*extChPacketConn).msgs
	msgs <- chPacketMsg{}
	go func() {
		c.Until(time.Second, func() bool { return len(msgs) == 0 })
		l0.Close()
	}()

	b := [8]byte{}
	_, _, err = l0.ReadFrom(b[:])
	c.MustError(err, "should be disconnected")
}

func TestChPacketWriteRemoteClosed(t *testing.T) {
	addr, c, nc := newChTest(t)

	l0, err := nc.listenPacket(addr + "0")
	c.MustNotError(err, "failed to listen")
	defer l0.Close()

	l1, err := nc.listenPacket(addr + "1")
	c.MustNotError(err, "failed to listen")
	defer l1.Close()

	msgs := l0.(*extChPacketConn).msgs
	for i := 0; i < cap(msgs); i++ {
		l1.WriteTo([]byte("test"), l0.LocalAddr())
	}

	go func() {
		time.Sleep(time.Millisecond)
		l0.Close()
	}()

	_, err = l1.WriteTo([]byte("test"), l0.LocalAddr())
	c.MustError(err, "should be disconnected")
}

func TestChPacketWriteLocalClosed(t *testing.T) {
	addr, c, nc := newChTest(t)

	l0, err := nc.listenPacket(addr + "0")
	c.MustNotError(err, "failed to listen")
	defer l0.Close()

	l1, err := nc.listenPacket(addr + "1")
	c.MustNotError(err, "failed to listen")
	defer l1.Close()

	msgs := l0.(*extChPacketConn).msgs
	for i := 0; i < cap(msgs); i++ {
		l1.WriteTo([]byte("test"), l0.LocalAddr())
	}

	go func() {
		time.Sleep(time.Millisecond)
		l1.Close()
	}()

	_, err = l1.WriteTo([]byte("test"), l0.LocalAddr())
	c.MustError(err, "should be disconnected")
}

func TestChPacketDeadlines(t *testing.T) {
	addr, c, nc := newChTest(t)

	l0, err := nc.listenPacket(addr + "0")
	c.MustNotError(err, "failed to listen")
	defer l0.Close()

	l1, err := nc.listenPacket(addr + "1")
	c.MustNotError(err, "failed to listen")
	defer l1.Close()

	msgs := l0.(*extChPacketConn).msgs
	for i := 0; i < cap(msgs); i++ {
		l1.WriteTo([]byte("test"), l0.LocalAddr())
	}

	l1.SetDeadline(time.Now().Add(time.Millisecond))
	_, err = l1.WriteTo([]byte("test"), l0.LocalAddr())
	c.MustError(err, "deadline should be hit")
	c.True(err.(net.Error).Timeout(), "should time out")

	b := [8]byte{}
	l1.SetDeadline(time.Now().Add(time.Millisecond))
	_, _, err = l1.ReadFrom(b[:])
	c.MustError(err, "deadline should be hit")
	c.True(err.(net.Error).Timeout(), "should time out")
	c.False(err.(net.Error).Temporary(), "should not be temporary")
	c.Equal(err.Error(), "deadline passed")
}
