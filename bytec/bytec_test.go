package bytec

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestDup(t *testing.T) {
	c := check.New(t)

	in := []byte("test")
	out := Dup(in)

	c.True(&in[0] != &out[0])
	c.Equal(in, out)
}
