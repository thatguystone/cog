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

func TestCheckNamePath(t *testing.T) {
	c := New(t)

	c.Equal("TestCheckNamePath", c.Name())
	c.Contains(c.Path(), "cog/check")
}

func TestCheckB(t *testing.T) {
	New(t)
	New(&testing.B{}).B()
}

func TestRunCoverage(t *testing.T) {
	c := New(t)

	c.Panics(func() {
		c := C{}
		c.Run("", nil)
	})

}

func TestGetTestNameCoverage(t *testing.T) {
	c := New(t)

	c.Panics(func() {
		getTestName(nil)
	})
}

func BenchmarkTest(b *testing.B) {
	New(b)

	b.Run("A", func(b *testing.B) {
		New(b)
	})

	b.Run("A", func(b *testing.B) {
		New(b)
	})
}
