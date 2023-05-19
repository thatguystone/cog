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

func TestGet(t *testing.T) {
	c := check.NewT(t)
	funcName := pkgName + ".TestGet"

	st := Get()
	c.Equal(st.MakeFrames()[0].Function, funcName)
	c.True(strings.Contains(st.String(), funcName))

	const depth = 129
	expectDepth := len(st.MakeFrames()) + depth

	var fn func(n int)
	fn = func(n int) {
		if n != 1 {
			fn(n - 1)
			return
		}

		st := Get()
		frames := st.MakeFrames()

		c.Equalf(len(frames), expectDepth, "%s", st)
		c.Equal(frames[depth].Function, funcName)
	}

	fn(depth)
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
	for i := 0; i < 5; i++ {
		st, found := FromError(err)
		c.True(found)
		c.Equal(deep.Stack, st)

		err = fmt.Errorf("%s: %w", "wrap", deep)
	}

	_, found := FromError(errors.New("merp"))
	c.False(found)
}

func TestString(t *testing.T) {
	c := check.NewT(t)
	c.Equal(Stack{}.String(), "")
}

func BenchmarkGetSkip(b *testing.B) {
	c := check.NewB(b)

	const depth = 32
	expectDepth := len(Get().MakeFrames()) + depth

	var fn func(n int)
	fn = func(n int) {
		if n != 1 {
			fn(n - 1)
			return
		}

		st := Get()
		c.Must.Equalf(len(st.MakeFrames()), expectDepth, "%s", st)
		c.ResetTimer()

		for i := 0; i < c.N; i++ {
			Get().Iter(func(f Frame) {})
		}
	}

	fn(depth)
}
