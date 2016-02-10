package unsafec

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestString(t *testing.T) {
	c := check.New(t)

	b := []byte("string")
	us := String(b)
	c.Equal(string(b), us)
}

func TestBytes(t *testing.T) {
	c := check.New(t)

	s := "bytes"
	ub := Bytes(s)
	c.Equal([]byte(s), ub)
}
