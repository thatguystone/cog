package callstack

import (
	"errors"
	"fmt"
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
	c.Equal(st.MakeFrames()[0].Function, funcName)
	c.True(strings.Contains(st.String(), funcName))

	const depth = 129
	expectDepth := len(st.MakeFrames()) + depth

	frames := recurse(depth, Get).MakeFrames()
	c.Equalf(len(frames), expectDepth, "%s", st)
	c.Equal(frames[depth].Function, funcName)
}

func TestFromError(t *testing.T) {
	c := check.NewT(t)

	type embed struct {
		Stack
		error
	}

	deep := embed{
		Stack: Get(),
		error: errors.New("test"),
	}

	err := fmt.Errorf("%s: %w", "wrap", deep)
	for range 5 {
		st, found := FromError(err)
		c.True(found)
		c.Equal(deep.Stack, st)

		err = fmt.Errorf("%s: %w", "wrap", deep)
	}

	_, found := FromError(errors.New("merp"))
	c.False(found)
}

func TestStackIter(t *testing.T) {
	c := check.NewT(t)

	calls := 0
	recurse(10, Get).Iter(func(f Frame) bool {
		calls++
		return calls <= 2
	})

	c.Equal(calls, 3)
}

func TestStackString(t *testing.T) {
	c := check.NewT(t)
	c.Equal(Stack{}.String(), "")
}

func BenchmarkGetSkip(b *testing.B) {
	c := check.NewB(b)

	recurse(32, func() any {
		c.ResetTimer()

		for range b.N {
			Get().Iter(func(f Frame) bool { return true })
		}

		return nil
	})
}
