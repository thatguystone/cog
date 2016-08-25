package bytec

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestSlice(t *testing.T) {
	n := testing.AllocsPerRun(5000, func() {
		s := Make(123)
		*s = append(*s, "1234"...)
		Put(&s)
	})

	c := check.New(t)
	c.Equal(n, 0.0)
}

func TestLog2(t *testing.T) {
	c := check.New(t)

	c.Equal(log2(0), 0)
	c.Equal(log2f(0), 0)

	for i := uint64(1); i != 0; i *= 2 {
		l := log2(uint64(i))
		m := log2f(uint64(i))
		c.Equal(l, m, "at %d")
	}
}

func TestPoolN(t *testing.T) {
	c := check.New(t)

	c.Equal(poolN(5, false), 0)
	c.Equal(poolN(10, false), 0)
	c.Equal(poolN(sPoolsStart-10, false), 0)
	c.Equal(poolN(sPoolsStart+10, false), 1)

	size := sPoolsStart
	for i := 0; i < sPoolsN*2; i++ {
		p := i + 1
		if p >= sPoolsN {
			p = sPoolsN - 1
		}

		c.Equal(p, poolN(size, false), "size=%d", size)
		size *= 2
	}
}

func BenchmarkPoolN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		size := sPoolsStart

		for i := 0; i < sPoolsN; i++ {
			poolN(size, false)
			size *= 2
		}
	}
}

func BenchmarkSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		size := sPoolsStart
		for i := 0; i < sPoolsN; i++ {
			s := Make(size)
			Put(&s)

			size *= 2
		}
	}
}
