package cnet

import "sync"

type ch struct {
	mtx         sync.Mutex
	connID      uint64
	listeners   map[string]*chListener
	packetConns map[string]*chPacketConn
}

type chAddr string

// ChError is the error returned from any operations that happen on channel-
// based sockets.
type ChError struct {
	msg     string
	timeout bool
	temp    bool
}

// Where all ch:// calls go
var globalChs = newCh()

func newCh() *ch {
	return &ch{
		listeners:   map[string]*chListener{},
		packetConns: map[string]*chPacketConn{},
	}
}

func (na chAddr) Network() string {
	return "ch"
}

func (na chAddr) String() string {
	return string(na)
}

func newChError(timeout, temp bool, msg string) ChError {
	return ChError{
		msg:     msg,
		timeout: timeout,
		temp:    temp,
	}
}

// Error returns the error message
func (e ChError) Error() string {
	return e.msg
}

//Timeout returns if this error is because of a timeout
func (e ChError) Timeout() bool {
	return e.timeout
}

// Temporary returns if this error is temporary, so that you can retry the
// operation
func (e ChError) Temporary() bool {
	return e.temp
}
