package statc

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func (snap *Snapshot) addTestData() {
	snap.Add(newName("int"), int64(1))
	snap.Add(newName("str"), "string")
	snap.Add(newName("bool.true"), true)
	snap.Add(newName("bool.false"), false)
	snap.Add(newName("float"), 1.2445)
}

func TestNameError(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		Name{}.Str()
	})
}

func BenchmarkSnapAdd(b *testing.B) {
	snap := Snapshot{}

	for i := 0; i < b.N; i++ {
		snap = snap[:0]
		snap.addTestData()
	}
}
