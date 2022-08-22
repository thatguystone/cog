package check

import (
	"testing"
)

func TestCheckT(t *testing.T) {
	c := NewT(t)
	c.Equal(1, 1)
	c.Run("test", func(c *T) {})
}

func FuzzCheckF(f *testing.F) {
	c := NewF(f)
	c.Equal(1, 1)

	c.Panics(func() {
		c.Fuzz(1)
	})

	c.Panics(func() {
		c.Fuzz(func() {})
	})

	c.Panics(func() {
		c.Fuzz(func(c *T, a int) bool { return true })
	})

	c.Add(1)
	c.Fuzz(func(c *T, a int) {
		c.Equal(a, 1)
	})
}

func BenchmarkTest(b *testing.B) {
	c := NewB(b)
	c.Run("A", func(c *B) {})
}
