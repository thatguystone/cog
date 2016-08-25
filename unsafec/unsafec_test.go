package unsafec

import (
	"testing"

	"github.com/iheartradio/cog/check"
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

func TestBytesCap(t *testing.T) {
	c := check.New(t)

	s := "bytes"
	ub := Bytes(s)

	c.Equal([]byte(s), ub)
	c.Equal(len(ub), cap(ub))
}

func BenchmarkString(b *testing.B) {
	b.ReportAllocs()

	s := []byte("bytes")
	for i := 0; i < b.N; i++ {
		String(s)
	}
}

func BenchmarkBytes(b *testing.B) {
	b.ReportAllocs()

	s := "bytes"
	for i := 0; i < b.N; i++ {
		Bytes(s)
	}
}
