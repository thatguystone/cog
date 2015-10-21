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
