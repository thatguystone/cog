package statc

import "testing"

func (snap *Snapshot) addTestData() {
	snap.Add("int", int64(1))
	snap.Add("str", "string")
	snap.Add("bool.true", true)
	snap.Add("bool.false", false)
	snap.Add("float", 1.2445)
}

func BenchmarkSnapAdd(b *testing.B) {
	snap := Snapshot{}

	for i := 0; i < b.N; i++ {
		snap = snap[:0]
		snap.addTestData()
	}
}
