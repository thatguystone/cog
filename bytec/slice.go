package bytec

import (
	"math"
	"sync"
)

// S is a []byte. Any existing []byte may be cast to an S and added to the pool.
type S []byte

const (
	sPoolsN     = 9
	sPoolsExp   = 7
	sPoolsStart = uint64(1 << sPoolsExp)
)

// The final pool is the overflow pool that just holds oversized buffers of no
// specific size.
var sPools = [sPoolsN]sync.Pool{}

func init() {
	curr := sPoolsStart
	for i := range sPools {
		size := curr
		curr *= 2

		sPools[i].New = func() interface{} {
			ss := make(S, 0, size)
			return &ss
		}
	}
}

// Used by everyone not on amd64
func log2f(n uint64) uint64 {
	if n == 0 {
		return 0
	}

	return uint64(math.Log2(float64(n)))
}

func poolN(n uint64, cap bool) uint64 {
	i := log2(n)

	// If not looking for capacity, need to go to next slot: log2(156)==7, which
	// is the slot for cap==128. In this case, would need cap==256 to hold all
	// 156 bytes.
	if !cap {
		i++
	}

	if i < sPoolsExp {
		return 0
	}

	i -= sPoolsExp
	if i >= sPoolsN {
		return sPoolsN - 1
	}

	return i
}

// Make gets an appropriately-sized S from the pool. When you're done, be sure
// to call Put() on the S to add it back to the pool.
//
// When you request a buffer, the only guarauntee is that you will get a []byte
// _at least_ as large as the requested size, though likely larger.
func Make(n uint64) (s *S) {
	return sPools[poolN(n, false)].Get().(*S)
}

// Put returns the slice to the pool and sets S to nil.
func Put(s **S) {
	if *s != nil {
		n := poolN(uint64(cap(**s)), true)
		**s = (**s)[:0]

		sPools[n].Put(*s)
		*s = nil
	}
}
