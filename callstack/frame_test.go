package callstack

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

const pkgPath = "github.com/thatguystone/cog/callstack"

func TestSelfFunc(t *testing.T) {
	c := check.NewT(t)

	fr := Self()

	const funcName = "TestSelfFunc"
	c.Equal(fr.Function, pkgPath+"."+funcName)
	c.Equal(fr.FuncName(), funcName)
	c.Equal(fr.PkgPath(), pkgPath)
}

func TestSelfMethod(t *testing.T) {
	c := check.NewT(t)

	fr := testSelf{}.getFrame()

	const funcName = "testSelf.getFrame"
	c.Equal(fr.Function, pkgPath+"."+funcName)
	c.Equal(fr.FuncName(), funcName)
	c.Equal(fr.PkgPath(), pkgPath)
}

type testSelf struct{}

func (testSelf) getFrame() Frame {
	return Self()
}