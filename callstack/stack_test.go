package callstack

import (
	"reflect"
	"slices"
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
	check.Equal(t, slices.Collect(st.Frames())[0].Func(), funcName)
	check.True(t, strings.Contains(st.String(), funcName))

	const depth = 129
	expectDepth := len(slices.Collect(st.Frames())) + depth

	frames := slices.Collect(recurse(depth, Get).Frames())
	check.Equalf(t, len(frames), expectDepth, "%s", st)
	check.Equal(t, frames[depth].Func(), funcName)
}

func TestStackIters(t *testing.T) {
	recurse(10, func() any {
		for range Get().Frames() {
			break
		}

		return nil
	})
}

func TestStackString(t *testing.T) {
	var stack Stack
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
