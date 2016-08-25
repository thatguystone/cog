package cnet

import (
	"bytes"
	"io"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/check"
)

func newChTest(t *testing.T) (string, *check.C, *ch) {
	return check.GetTestName(), check.New(t), newCh()
}

func testChStreamListener(t *testing.T) (
	*check.C,
	*ch,
	string,
	net.Listener,
	chan net.Conn,
	func()) {

	addr, c, nc := newChTest(t)

	l, err := nc.listen(addr)
	c.MustNotError(err, "failed to listen")

	accepts := make(chan net.Conn)
	closeCh := make(chan struct{})
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			select {
			case accepts <- c:
			case <-closeCh:
			}
		}
	}()

	cleanup := func() {
		l.Close()
		close(closeCh)
	}

	return c, nc, addr, l, accepts, cleanup
}

func netChConnPair(c *check.C, nc *ch, addr string, accepts chan net.Conn) (
	net.Conn,
	net.Conn) {

	cs, err := nc.dial("test", addr, time.Second)
	c.MustNotError(err, "failed to dial")

	cr := <-accepts

	return cs, cr
}

func TestChStreamStreamBasic(t *testing.T) {
	c := check.New(t)
	n := New(check.GetTestName())

	addr := "ch://" + check.GetTestName()
	l, err := n.Listen(addr)
	c.MustNotError(err)
	defer l.Close()

	c.Equal(check.GetTestName(), l.Addr().String())

	_, err = n.Dial(addr, time.Millisecond)
	c.MustError(err)
	c.True(err.(ChError).Timeout())
	c.True(!err.(ChError).Temporary())

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			go echoConn(c)
		}
	}()

	conn, err := n.Dial(addr, time.Second)
	c.MustNotError(err)
	defer conn.Close()

	payload := []byte("Test")
	sent, err := conn.Write(payload)
	c.MustNotError(err)
	c.Equal(4, sent)

	buff := make([]byte, 128)
	got, err := conn.Read(buff)
	c.MustNotError(err)
	c.Equal(4, sent)
	c.Equal(payload, buff[0:got])
}

func TestChStreamStreamGCListen(t *testing.T) {
	c := check.New(t)
	nc := newCh()

	addr := "ch://one"
	_, err := nc.listen(addr)
	c.MustNotError(err)

	c.Until(time.Second, func() bool {
		runtime.GC()
		return len(nc.listeners) == 0
	})
}

func TestChStreamStreamGCConn(t *testing.T) {
	c := check.New(t)
	nc := newCh()

	addr := "one"
	l, err := nc.listen(addr)
	c.MustNotError(err)
	defer l.Close()

	accepts := make(chan net.Conn)
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			accepts <- c
		}
	}()

	conn, err := nc.dial("one", addr, time.Second)
	c.MustNotError(err)

	<-accepts
	c.Until(time.Second, func() bool {
		runtime.GC()
		return *(conn.(*chConn)).closed
	})

	_, err = nc.dial("one", addr, time.Second)
	c.MustNotError(err)

	conn = <-accepts
	c.Until(time.Second, func() bool {
		runtime.GC()
		return *(conn.(*chConn)).closed
	})
}

func TestChStreamListenAddressInUse(t *testing.T) {
	c := check.New(t)
	nc := newCh()

	l, err := nc.listen("one")
	c.MustNotError(err)
	defer l.Close()

	_, err = nc.listen("one")
	c.MustError(err)
	c.True(!err.(ChError).Timeout())
	c.True(!err.(ChError).Temporary())
}

func TestChStreamDialConnRefused(t *testing.T) {
	c := check.New(t)
	nc := newCh()

	_, err := nc.dial("something", "one", time.Second)
	c.MustError(err)
	c.True(!err.(ChError).Timeout())
	c.True(!err.(ChError).Temporary())
}

func TestChStreamReadWriteClosed(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)
	cs.Close()
	cr.Close()

	b := [8]byte{}
	_, err := cs.Read(b[:])
	c.Error(err, "read should fail when closed")

	_, err = cr.Read(b[:])
	c.Error(err, "read should fail when closed")

	_, err = cs.Write(b[:])
	c.Error(err, "write should fail when closed")

	_, err = cr.Write(b[:])
	c.Error(err, "write should fail when closed")
}

func TestChStreamReadClosed(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)

	b := []byte("after close")
	_, err := cs.Write(b)
	c.MustNotError(err)
	cs.Close()

	buff := make([]byte, 1)
	for _, b := range b {
		n, err := cr.Read(buff)
		c.MustEqual(n, 1)
		c.MustNotError(err)
		c.MustEqual(buff[0], b)
	}

	n, err := cr.Read(buff)
	c.MustEqual(n, 0)
	c.Equal(err, io.EOF)
}

func TestChStreamWriteClosed(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)

	timer := time.NewTimer(100 * time.Millisecond)
	go func() {
		<-timer.C
		cr.Close()
	}()

	var err error
	for err == nil {
		timer.Reset(100 * time.Millisecond)
		_, err = cs.Write([]byte{})
	}

	c.Equal(io.ErrClosedPipe, err)
}

func TestChStreamConnAddrs(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)
	defer cs.Close()
	defer cr.Close()

	c.Equal("ch", cs.RemoteAddr().Network())
	c.Equal("TestChStreamConnAddrs", cs.RemoteAddr().String())
	c.Equal(cs.LocalAddr(), cr.RemoteAddr(), "addresses should cross match")
	c.Equal(cr.LocalAddr(), cs.RemoteAddr(), "addresses should cross match")
}

func TestChStreamConnCloseInRead(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)
	defer cs.Close()
	defer cr.Close()

	inCh := cs.(*chConn).inCh
	cog.Notify(inCh)
	go func() {
		c.Until(time.Second, func() bool { return len(inCh) == 0 })
		cs.Close()
	}()

	buff := [8]byte{}
	cs.Read(buff[:])

	// No more checks here, Read() should return and this should be done
}

func TestChStreamConnNotification(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)
	defer cs.Close()
	defer cr.Close()

	gotRead := make(chan struct{})

	go func() {
		b := [4]byte{}
		for i := 0; i < 5; i++ {
			_, err := cr.Read(b[:])
			c.MustNotError(err, "read should not fail")
			gotRead <- struct{}{}
		}
	}()

	for i := 0; i < 5; i++ {
		_, err := cs.Write([]byte("test"))
		c.MustNotError(err, "write should not fail")
		<-gotRead
	}
}

func TestChStreamConnPartialRead(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)
	defer cs.Close()
	defer cr.Close()

	cs.Write(bytes.Repeat([]byte("test"), 100))
	for i := 0; i < 10; i++ {
		in := [8]byte{}
		l, err := cr.Read(in[:])
		c.Equal(len(in), l, "did not fill buffer")
		c.Equal("testtest", string(in[:]), "did not get test string")
		c.Equal(err, nil, "should not get an error for read")
	}
}

func TestChStreamConnDeadlines(t *testing.T) {
	c, nc, addr, _, accepts, cleanup := testChStreamListener(t)
	defer cleanup()

	cs, cr := netChConnPair(c, nc, addr, accepts)
	defer cs.Close()
	defer cr.Close()

	cs.SetDeadline(time.Now().Add(time.Millisecond))

	b := [8]byte{}
	_, err := cs.Read(b[:])
	c.MustError(err, "should get error")
	c.True(err.(net.Error).Timeout(), "should time out")
}

func TestChStreamDialInvalidAddress(t *testing.T) {
	_, c, nc := newChTest(t)

	_, err := nc.dial("test", "ch://this does not exist", time.Nanosecond)
	c.MustError(err, "dial didn't fail")
}

func TestChStreamDialTimeout(t *testing.T) {
	addr, c, nc := newChTest(t)

	l, err := nc.listen(addr)
	c.MustNotError(err, "failed to listen")
	defer l.Close()

	_, err = nc.dial("test", addr, time.Nanosecond)
	c.MustError(err, "dial didn't fail")
	c.True(err.(net.Error).Timeout(), "should fail with timeout")
}

func TestChStreamListenerClose(t *testing.T) {
	addr, c, nc := newChTest(t)

	l, err := nc.listen(addr)
	c.MustNotError(err, "failed to listen")
	defer l.Close()

	go func() {
		c.Until(time.Second, func() bool {
			return l.(*extChListener).pendingAccepts > 0
		})
		l.Close()

		_, err := l.Accept()
		c.Error(err, "should error after when closed")
	}()

	_, err = nc.dial("test", addr, time.Second)
	c.MustError(err, "dial didn't fail")
	c.False(err.(net.Error).Timeout(),
		"should fail with conn refused, not timeout")
}
