package callstack

import (
	"fmt"
	"runtime"
	"strings"
)

// A Stack is an immutable stack trace
type Stack struct {
	s []uintptr
}

// Get gets a stack trace at the caller's location
func Get() Stack {
	return GetSkip(1)
}

// GetSkip gets the stack, skipping the first skip number of frames
func GetSkip(skip int) Stack {
	// Exclude self and runtime.Callers()
	skip += 2

	pcs := make([]uintptr, 64)
	for {
		n := runtime.Callers(skip, pcs)
		if n < len(pcs) {
			return Stack{
				s: pcs[:n],
			}
		}

		pcs = append(pcs, make([]uintptr, len(pcs))...)
	}
}

// A shallow wrapper around [runtime.CallersFrames]
func (st Stack) CallersFrames() *runtime.Frames {
	return runtime.CallersFrames(st.s)
}

// Iter calls cb for each frame in the [Stack]
func (st Stack) Iter(cb func(runtime.Frame)) {
	frames := st.CallersFrames()
	for {
		frame, more := frames.Next()
		if frame != (runtime.Frame{}) {
			cb(frame)
		}

		if !more {
			return
		}
	}
}

// MakeFrames expands the stack into a slice of [runtime.Frame]
func (st Stack) MakeFrames() (ret []runtime.Frame) {
	ret = make([]runtime.Frame, 0, len(st.s))

	st.Iter(func(f runtime.Frame) {
		ret = append(ret, f)
	})

	return
}

// String implements [fmt.Stringer]
func (st Stack) String() string {
	var b strings.Builder

	st.Iter(func(f runtime.Frame) {
		fmt.Fprintf(&b, "%s()\n", f.Function)
		fmt.Fprintf(&b, "\t%s:%d\n", f.File, f.Line)
	})

	return b.String()
}
