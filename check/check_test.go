package check

import "testing"

func TestMain(m *testing.M) {
	Main(m)
}

func TestCheckT(t *testing.T) {
	New(t).T()
}

func TestCheckB(t *testing.T) {
	New(t)
	New(&testing.B{}).B()
}

func TestCheckMultipleNew(t *testing.T) {
	for i := 0; i < 5; i++ {
		New(t)
	}
}

func TestCheckNewError(t *testing.T) {
	c := New(t)

	c.Panics(func() {
		New(new(testing.T))
	})
}

func TestGetTestName(t *testing.T) {
	c := New(t)
	c.Equal("TestGetTestName", GetTestName())

	func() {
		c.Equal("TestGetTestName", GetTestName())
	}()
}
