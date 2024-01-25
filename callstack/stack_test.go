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
	c.Equal(st.Slice()[0].Function, funcName)
	c.True(strings.Contains(st.String(), funcName))

	const depth = 129
	expectDepth := len(st.Slice()) + depth

	frames := recurse(depth, Get).Slice()
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
	c.Equal(Stack{}.String(), "")
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
