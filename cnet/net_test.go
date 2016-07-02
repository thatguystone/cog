package cnet

import (
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func newTest(t *testing.T) (*check.C, Net) {
	c := check.New(t)
	return c, New(c.Name())
}

func TestNetBasic(t *testing.T) {
	c, net := newTest(t)

	_, err := net.Dial("tcp:/123414124", time.Second)
	c.Must.NotNil(err)
}

func TestHostExists(t *testing.T) {
	c, net := newTest(t)

	c.True(net.HostExists("ch://" + c.Name()))
	c.True(net.HostExists("localhost:123"))
	c.True(net.HostExists("localhost"))
}

func TestResolve(t *testing.T) {
	c, net := newTest(t)

	addrs := []string{
		"tcp://localhost:123",
		"udp://localhost:123",
		"ch://localhost",
	}

	for _, addr := range addrs {
		_, err := net.Resolve("tcp", addr)
		c.Nil(err, "failed at %s", addr)
	}

	_, err := net.Resolve("", "merp://asdasd")
	c.NotNil(err)
}

func TestListen(t *testing.T) {
	c, net := newTest(t)

	_, err := net.Listen(":0")
	c.Nil(err)
}

func TestListenPacket(t *testing.T) {
	c, net := newTest(t)

	_, err := net.ListenPacket("ch://test")
	c.Nil(err)

	_, err = net.ListenPacket(":0")
	c.Nil(err)
}
