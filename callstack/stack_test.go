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
	funcName := pkgName + ".TestGet"

	st := Get()
	check.Equal(t, st.Slice()[0].Func(), funcName)
	check.True(t, strings.Contains(st.String(), funcName))

	const depth = 129
	expectDepth := len(st.Slice()) + depth

	frames := recurse(depth, Get).Slice()
	check.Equalf(t, len(frames), expectDepth, "%s", st)
	check.Equal(t, frames[depth].Func(), funcName)
}

func TestStackIters(t *testing.T) {
	recurse(10, func() any {
		for range Get().All() {
			break
		}

		return nil
	})
}

func TestStackString(t *testing.T) {
	var stack Stack
	check.True(t, stack.IsZero())
	check.Equal(t, stack.String(), "")
}

func BenchmarkGet(b *testing.B) {
	b.ReportAllocs()

	recurse(32, func() any {
		for b.Loop() {
			Get()
		}

		return nil
	})
}
