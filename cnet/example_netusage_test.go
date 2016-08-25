package cnet_test

import (
	"fmt"
	"net"
	"time"

	"github.com/iheartradio/cog/cnet"
)

func echoServer(c net.Conn) {
	buff := make([]byte, 1024)
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

func Example_netUsage() {
	net := cnet.New("NetUsageExample")

	// Create a new listener in process-local, channel-based socket space
	addr := "ch://some-channel-name"

	l, err := net.Listen(addr)
	if err != nil {
		panic(err)
	}

	defer l.Close()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			go echoServer(c)
		}
	}()

	c, err := net.Dial(addr, time.Second)
	if err != nil {
		panic(err)
	}

	defer c.Close()

	_, err = c.Write([]byte("test message"))
	if err != nil {
		panic(err)
	}

	buff := make([]byte, 128)
	n, err := c.Read(buff)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(buff[0:n]))

	// Output:
	// test message
}
