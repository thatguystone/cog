package cnet

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestRemoveDefaultPort(t *testing.T) {
	c := check.New(t)

	tests := []struct {
		in   string
		port string
		out  string
	}{
		{
			in:   "test:8080",
			port: "8080",
			out:  "test",
		},
		{
			in:   "test",
			port: "8080",
			out:  "test",
		},
		{
			in:   ":80",
			port: "80",
			out:  "",
		},
	}

	for _, test := range tests {
		out := RemoveDefaultPort(test.in, test.port)
		c.Equal(test.out, out)
	}
}
