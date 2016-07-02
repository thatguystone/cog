package cnet

import (
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func newXTest(t *testing.T) (*check.C, XNet) {
	c := check.New(t)
	return c, NewX(c.Name())
}

func echoConn(c net.Conn) {
	buff := make([]byte, 128)
	for {
		got, err := c.Read(buff)
		if err == nil {
			_, err = c.Write(buff[0:got])
		}

		if err != nil {
			return
		}
	}
}

func TestXNetBasic(t *testing.T) {
	c, x := newXTest(t)

	x.SetOffline(true)

	addr := "ch://" + c.Name()
	_, err := x.Listen(addr)
	c.Must.Nil(err)

	_, err = x.Dial(addr, time.Second)
	c.Must.NotNil(err)
}

func TestXNetConnGC(t *testing.T) {
	c, x := newXTest(t)

	addr := "ch://" + c.Name()
	l, err := x.Listen(addr)
	c.Must.Nil(err)
	defer l.Close()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			go echoConn(c)
		}
	}()

	conn, err := x.Dial(addr, time.Second)
	c.Must.Nil(err)

	cc := conn.(*xConn).Conn.(*chConn)
	c.Until(time.Second, func() bool {
		runtime.GC()
		return *cc.closed
	})
}

func TestXNetAcceptOffline(t *testing.T) {
	c, x := newXTest(t)
	x2 := NewX(c.Name() + "2")

	x.SetOffline(true)

	l, err := x.Listen("127.0.0.1:0")
	c.Must.Nil(err)
	defer l.Close()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			go echoConn(c)
		}
	}()

	// Connection should succeed but be immediately closed
	conn, err := x2.Dial(l.Addr().String(), time.Second)
	c.Must.Nil(err)

	buff := make([]byte, 4)
	n := -1
	for n != 0 {
		n, _ = conn.Read(buff)
	}
}

func TestXNetCloseOnSetOffline(t *testing.T) {
	c, x := newXTest(t)

	addr := "ch://" + c.Name()
	l, err := x.Listen(addr)
	c.Must.Nil(err)
	defer l.Close()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			go echoConn(c)
		}
	}()

	conn, err := x.Dial(addr, time.Second)
	c.Must.Nil(err)

	_, err = conn.Write([]byte("a"))
	c.Must.Nil(err)

	buff := make([]byte, 4)
	n, err := conn.Read(buff)
	c.Must.Nil(err)
	c.Equal(1, n)

	x.SetOffline(true)

	_, err = conn.Write([]byte("a"))
	c.Must.NotNil(err)
}

func TestXNetPacket(t *testing.T) {
	c, x := newXTest(t)

	addr := "ch://" + c.Name()
	l, err := x.ListenPacket(addr)
	c.Must.Nil(err)

	to := chAddr(c.Name())
	n, err := l.WriteTo([]byte("Test"), to)
	c.Equal(4, n)
	c.Must.Nil(err)

	buff := make([]byte, 4)
	n, from, err := l.ReadFrom(buff)
	c.Equal(4, n)
	c.Equal(to, from)
	c.Nil(err)
}

func TestXNetPacketOffline(t *testing.T) {
	c, x := newXTest(t)

	x.SetOffline(true)

	addr := "ch://" + c.Name()
	l, err := x.ListenPacket(addr)
	c.Must.Nil(err)

	n, err := l.WriteTo([]byte("Test"), chAddr(c.Name()))
	c.Equal(4, n)
	c.Must.Nil(err)

	buff := make([]byte, 4)
	l.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	n, raddr, err := l.ReadFrom(buff)
	c.Equal(0, n)
	c.Equal(nil, raddr)
	c.True(err.(ChError).Timeout())
}

func TestXNetOfflineErrors(t *testing.T) {
	c, x := newXTest(t)

	c.True(x.HostExists("localhost"))
	_, err := x.Resolve("tcp", "localhost:123")
	c.Nil(err)

	x.SetOffline(true)

	c.False(x.HostExists("localhost"))
	_, err = x.Resolve("tcp", "localhost:123")
	c.Equal(ErrOffline, err)
}
