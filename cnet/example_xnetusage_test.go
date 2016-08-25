package cnet_test

import (
	"errors"
	"time"

	"github.com/iheartradio/cog/cnet"
)

func Example_xNetUsage() {
	xnet := cnet.NewX("XNetUsageExample")

	// Don't let anything flow.
	xnet.SetOffline(true)

	l, err := xnet.Listen("127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	defer l.Close()

	go func() {
		for {
			_, err := l.Accept()
			if err != nil {
				return
			}
		}
	}()

	_, err = xnet.Dial(l.Addr().String(), time.Second)
	if err == nil {
		panic(errors.New("xnet is offline, Dial should fail"))
	}

	// Allow stuff to start flowing again
	xnet.SetOffline(false)

	// Dial should succeed now: since everything is back online, the connection
	// will be allowed through.
	_, err = xnet.Dial(l.Addr().String(), time.Second)
	if err != nil {
		panic(err)
	}

	// Output:
}
