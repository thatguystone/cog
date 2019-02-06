package check

import "testing"

func TestCheckT(t *testing.T) {
	New(t).T()

	t.Run("A", func(t *testing.T) {
		New(t)
	})

	t.Run("A", func(t *testing.T) {
		New(t)
	})

	t.Run("A", func(t *testing.T) {
		New(t)
	})
}

func TestCheckB(t *testing.T) {
	New(new(testing.B)).B()
}

func TestCheckRun(t *testing.T) {
	c := New(t)

	c.Run("derp", func(c *C) {
		c.Equal(1, 1)
	})
}

func TestCheckRunCoverage(t *testing.T) {
	c := New(t)

	c.Panics(func() {
		c := C{}
		c.Run("", nil)
	})
}

func BenchmarkTest(b *testing.B) {
	c := New(b)
	c.Run("A", func(c *C) {})
}
