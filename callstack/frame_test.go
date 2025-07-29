package callstack

import (
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

const (
	pkgPath  = "github.com/thatguystone/cog/callstack"
	fileName = "frame_test.go"
)

func TestSelfFunc(t *testing.T) {
	fr := Self().Frame()

	const funcName = "TestSelfFunc"
	check.NotEqual(t, fr.PC(), uintptr(0))
	check.Equal(t, fr.PkgPath(), pkgPath)
	check.Equal(t, fr.Func(), pkgPath+"."+funcName)
	check.Equal(t, fr.FuncName(), funcName)
	check.True(t, strings.Contains(fr.File(), fileName))
	check.Equal(t, fr.FileName(), fileName)
	check.NotEqual(t, fr.Line(), 0)
}

func TestSelfMethod(t *testing.T) {
	fr := testSelf{}.getPC().Frame()

	const funcName = "testSelf.getPC"
	check.NotEqual(t, fr.PC(), uintptr(0))
	check.Equal(t, fr.PkgPath(), pkgPath)
	check.Equal(t, fr.Func(), pkgPath+"."+funcName)
	check.Equal(t, fr.FuncName(), funcName)
	check.True(t, strings.Contains(fr.File(), fileName))
	check.Equal(t, fr.FileName(), fileName)
	check.NotEqual(t, fr.Line(), 0)
}

func TestPCZero(t *testing.T) {
	var pc PC
	fr := pc.Frame()
	check.Equal(t, fr.PkgPath(), "???")
	check.Equal(t, fr.Func(), "???")
	check.Equal(t, fr.FuncName(), "???")
	check.Equal(t, fr.File(), "???")
	check.Equal(t, fr.FileName(), "???")
	check.Equal(t, fr.Line(), 0)
}

func TestFrameString(t *testing.T) {
	str := Self().Frame().String()
	check.True(t, strings.Contains(str, fileName))
}

type testSelf struct{}

func (testSelf) getPC() PC {
	return Self()
}

func BenchmarkSelf(b *testing.B) {
	b.ReportAllocs()

	recurse(10, func() any {
		for b.Loop() {
			Self()
		}

		return nil
	})
}
