package cnet

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type chListener struct {
	mtx            sync.Mutex
	ch             *ch
	addr           chAddr
	pendingAccepts int32
	acceptCh       chan *chConn
	closed         bool
	closeCh        chan struct{}
}

type extChListener struct {
	*chListener
}

type chConn struct {
	mtx          *sync.Mutex
	ch           *ch
	localAddr    chAddr
	remoteAddr   chAddr
	readDeadline time.Time
	in           *bytes.Buffer
	inCh         chan struct{}
	out          *bytes.Buffer
	outCh        chan struct{}
	closed       *bool
	closeCh      chan struct{}
}

func (nc *ch) listen(addr string) (net.Listener, error) {
	nc.mtx.Lock()
	defer nc.mtx.Unlock()

	_, ok := nc.listeners[addr]
	if ok {
		return nil, newChError(
			false, false,
			fmt.Sprintf("tcp address in use: %s", addr))
	}

	l := &chListener{
		ch:       nc,
		addr:     chAddr(addr),
		acceptCh: make(chan *chConn),
		closeCh:  make(chan struct{}),
	}

	el := &extChListener{l}
	nc.listeners[addr] = l

	runtime.SetFinalizer(el, finalizeChListener)

	return el, nil
}

func finalizeChListener(el *extChListener) {
	el.Close()
}

func (nc *ch) dial(from, addr string, t time.Duration) (net.Conn, error) {
	cid := atomic.AddUint64(&nc.connID, 1)
	closed := false
	conn := &chConn{
		mtx:        &sync.Mutex{},
		ch:         nc,
		localAddr:  chAddr(fmt.Sprintf("%s#%d", from, cid)),
		remoteAddr: chAddr(addr),
		in:         &bytes.Buffer{},
		inCh:       make(chan struct{}, 1),
		out:        &bytes.Buffer{},
		outCh:      make(chan struct{}, 1),
		closed:     &closed,
		closeCh:    make(chan struct{}),
	}

	nc.mtx.Lock()
	l, ok := nc.listeners[addr]
	nc.mtx.Unlock()
	if !ok {
		return nil, newChError(false, false, "connection refused")
	}

	err := l.acceptConn(conn, t)
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(conn, finalizeChConn)

	return conn, nil
}

func finalizeChConn(c *chConn) {
	c.Close()
}

func (l *chListener) acceptConn(c *chConn, t time.Duration) error {
	atomic.AddInt32(&l.pendingAccepts, 1)
	defer atomic.AddInt32(&l.pendingAccepts, -1)

	select {
	case <-time.After(t):
		return newChError(
			true, false,
			fmt.Sprintf("timeout connecting to: ch://%s", l.addr))

	case <-l.closeCh:
		return newChError(false, false, "connection refused")

	case l.acceptCh <- c:
		return nil
	}
}

func (l *extChListener) Accept() (c net.Conn, err error) {
	err = func() error {
		l.mtx.Lock()
		defer l.mtx.Unlock()

		if l.closed {
			return newChError(false, false, "listener closed")
		}

		return nil
	}()

	if err != nil {
		return
	}

	select {
	case c := <-l.acceptCh:
		ret := &chConn{
			mtx:        c.mtx,
			ch:         c.ch,
			localAddr:  c.remoteAddr,
			remoteAddr: c.localAddr,
			in:         c.out,
			inCh:       c.outCh,
			out:        c.in,
			outCh:      c.inCh,
			closed:     c.closed,
			closeCh:    c.closeCh,
		}
		runtime.SetFinalizer(ret, finalizeChConn)
		return ret, nil

	case <-l.closeCh:
		return nil, newChError(false, false, "listener closed")
	}
}

func (l *extChListener) Addr() net.Addr {
	return l.addr
}

func (l *extChListener) Close() error {
	runtime.SetFinalizer(l, nil)

	func() {
		l.ch.mtx.Lock()
		defer l.ch.mtx.Unlock()

		delete(l.ch.listeners, l.addr.String())
	}()

	func() {
		l.mtx.Lock()
		defer l.mtx.Unlock()

		if !l.closed {
			l.closed = true
			close(l.closeCh)
		}
	}()

	return nil
}

func (c *chConn) readClosed(b []byte) (n int, err error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.in.Len() == 0 {
		err = io.EOF
		return
	}

	return c.in.Read(b)
}

func (c *chConn) Read(b []byte) (n int, err error) {
	inLen := func() (n int) {
		c.mtx.Lock()
		n = c.in.Len()
		c.mtx.Unlock()
		return
	}

	select {
	case <-c.closeCh:
		return c.readClosed(b)
	default:
		var timeout <-chan time.Time
		if !c.readDeadline.IsZero() {
			timeout = time.After(c.readDeadline.Sub(time.Now()))
		}

		for inLen() == 0 {
			select {
			case <-c.inCh:

			case <-timeout:
				err = newChError(true, false, "deadline passed")
				return

			case <-c.closeCh:
				return c.readClosed(b)
			}
		}

		c.mtx.Lock()
		defer c.mtx.Unlock()

		return c.in.Read(b)
	}
}

func (c *chConn) Write(b []byte) (n int, err error) {
	closed := func() bool {
		c.mtx.Lock()
		defer c.mtx.Unlock()

		closed := *c.closed

		if !closed {
			n, err = c.out.Write(b)
		}

		return closed
	}()

	if !closed {
		select {
		case <-c.closeCh:
			closed = true
		case c.outCh <- struct{}{}:
		}
	}

	if closed {
		return 0, io.ErrClosedPipe
	}

	return
}

func (c *chConn) Close() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if !*c.closed {
		*c.closed = true
		close(c.closeCh)
	}

	return nil
}

func (c *chConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *chConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *chConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *chConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *chConn) SetWriteDeadline(t time.Time) error {
	return nil
}
