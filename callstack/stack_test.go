package callstack

import (
	"reflect"
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

type self struct{}

var pkgName = reflect.TypeOf(self{}).PkgPath()

func recurse[T any](n int, cb func() T) T {
	if n > 1 {
		return recurse(n-1, cb)
	}

	return cb()
}

func TestGet(t *testing.T) {
	c := check.NewT(t)
	funcName := pkgName + ".TestGet"

	st := Get()
	c.Equal(st.Slice()[0].Func(), funcName)
	c.True(strings.Contains(st.String(), funcName))

	const depth = 129
	expectDepth := len(st.Slice()) + depth

	frames := recurse(depth, Get).Slice()
	c.Equalf(len(frames), expectDepth, "%s", st)
	c.Equal(frames[depth].Func(), funcName)
}

func TestStackIters(t *testing.T) {
	recurse(10, func() any {
		for _ = range Get().All() {
			break
		}

		return nil
	})
}

func TestStackString(t *testing.T) {
	c := check.NewT(t)

	var stack Stack
	c.True(stack.IsZero())
	c.Equal(stack.String(), "")
}

func BenchmarkGet(b *testing.B) {
	c := check.NewB(b)

	recurse(32, func() any {
		c.ResetTimer()

		for range b.N {
			Get()
		}

		return nil
	})
}
