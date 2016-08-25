package clog

import (
	"testing"

	"github.com/tchap/go-patricia/patricia"
	"github.com/iheartradio/cog/check"
)

func TestModulePrefix(t *testing.T) {
	c := check.New(t)

	tests := []struct {
		in   string
		pfx  patricia.Prefix
		name string
	}{
		{
			in:   "test",
			pfx:  patricia.Prefix("test."),
			name: "test",
		},
		{
			in:   "test.",
			pfx:  patricia.Prefix("test."),
			name: "test",
		},
		{
			in:   ".",
			pfx:  patricia.Prefix(""),
			name: "",
		},
		{
			in:   "....",
			pfx:  patricia.Prefix(""),
			name: "",
		},
		{
			in:   "",
			pfx:  patricia.Prefix(""),
			name: "",
		},
	}

	for _, test := range tests {
		pfx, name := modulePrefix(test.in)
		c.Equal(string(test.pfx), string(pfx))
		c.Equal(test.name, name)
	}
}
