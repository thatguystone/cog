package cort

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"testing"

	"github.com/iheartradio/cog/check"
)

type intSlice []int
type intpSlice []*int

func (p intSlice) Len() int           { return len(p) }
func (p intSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p intSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p intSlice) Move(i, j, a0, a1, b0, b1 int) {
	e := p[i]
	copy(p[a0:a1], p[b0:b1])
	p[j] = e
}

func (p intpSlice) Len() int           { return len(p) }
func (p intpSlice) Less(i, j int) bool { return *p[i] < *p[j] }
func (p intpSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p intpSlice) Move(i, j, a0, a1, b0, b1 int) {
	e := p[i]
	copy(p[a0:a1], p[b0:b1])
	p[j] = e
}

func makeIntSlice(n, max int) intSlice {
	v := intSlice{}
	for i := 0; i < n; i++ {
		v = append(v, rand.Intn(max))
	}

	sort.Sort(v)

	return v
}

func plop(v intSlice, max int) int {
	at := rand.Intn(len(v))
	v[at] = rand.Intn(max)
	return at
}

func TestFix(t *testing.T) {
	const max = 1000

	c := check.New(t)

	wg := sync.WaitGroup{}
	procs := runtime.GOMAXPROCS(-1)
	for i := 0; i < procs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			v := makeIntSlice(100, max)

			for i := 0; i < 3000; i++ {
				at := plop(v, max)

				s := make(intSlice, len(v))
				copy(s, v)

				sort.Sort(s)
				Fix(at, v)

				c.MustEqual(s, v)
			}
		}()
	}

	wg.Wait()
}

func TestFixOrdering(t *testing.T) {
	const n = 10

	c := check.New(t)

	s := intpSlice{}
	for i := 0; i < n; i++ {
		v := new(int)
		*v = i
		s = append(s, v)
	}

	i5 := s[5]
	*s[0] = *i5
	Fix(0, s)

	c.Equal(i5, s[5])
	c.Equal(*i5, *s[4])

	i4 := s[4]
	*s[0] = *i5
	Fix(0, s)

	c.Equal(i5, s[5])
	c.Equal(i4, s[4])
	c.Equal(*i5, *s[4])
	c.Equal(*i5, *s[3])
}

func BenchmarkResort(b *testing.B) {
	const max = 5000

	v := makeIntSlice(2500, max)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		plop(v, max)
		sort.Sort(v)
	}
}

func BenchmarkFix(b *testing.B) {
	const max = 5000

	v := makeIntSlice(2500, max)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		at := plop(v, max)
		Fix(at, v)
	}
}
