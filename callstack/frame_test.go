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
	c := check.NewT(t)

	fr := Self().Frame()

	const funcName = "TestSelfFunc"
	c.NotEqual(fr.PC(), 0)
	c.Equal(fr.PkgPath(), pkgPath)
	c.Equal(fr.Func(), pkgPath+"."+funcName)
	c.Equal(fr.FuncName(), funcName)
	c.True(strings.Contains(fr.File(), fileName))
	c.Equal(fr.FileName(), fileName)
	c.NotEqual(fr.Line(), 0)
}

func TestSelfMethod(t *testing.T) {
	c := check.NewT(t)

	fr := testSelf{}.getPC().Frame()

	const funcName = "testSelf.getPC"
	c.NotEqual(fr.PC(), 0)
	c.Equal(fr.PkgPath(), pkgPath)
	c.Equal(fr.Func(), pkgPath+"."+funcName)
	c.Equal(fr.FuncName(), funcName)
	c.True(strings.Contains(fr.File(), fileName))
	c.Equal(fr.FileName(), fileName)
	c.NotEqual(fr.Line(), 0)
}

func TestPCZero(t *testing.T) {
	c := check.NewT(t)

	var pc PC
	fr := pc.Frame()
	c.Equal(fr.PkgPath(), "???")
	c.Equal(fr.Func(), "???")
	c.Equal(fr.FuncName(), "???")
	c.Equal(fr.File(), "???")
	c.Equal(fr.FileName(), "???")
	c.Equal(fr.Line(), 0)
}

func TestFrameString(t *testing.T) {
	c := check.NewT(t)

	str := Self().Frame().String()
	c.True(strings.Contains(str, fileName))
}

type testSelf struct{}

func (testSelf) getPC() PC {
	return Self()
}

func BenchmarkSelf(b *testing.B) {
	c := check.NewB(b)

	recurse(10, func() any {
		c.ResetTimer()

		for range b.N {
			Self()
		}

		return nil
	})
}
