package callstack

import (
	"fmt"
	"iter"
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

// IsZero determines if this stack is a zero-value
func (st Stack) IsZero() bool {
	return len(st.s) == 0
}

// A shallow wrapper around [runtime.CallersFrames]
func (st Stack) CallersFrames() *runtime.Frames {
	return runtime.CallersFrames(st.s)
}

func (st Stack) Slice() []Frame {
	ret := make([]Frame, 0, len(st.s))
	for f := range st.All() {
		ret = append(ret, f)
	}

	return ret
}

// All returns an iterator over every frame in the stack.
func (st Stack) All() iter.Seq[Frame] {
	return func(yield func(Frame) bool) {
		frames := st.CallersFrames()
		for {
			frame, more := frames.Next()
			if frame != (runtime.Frame{}) {
				if !yield(Frame{frame}) {
					return
				}
			}

			if !more {
				return
			}
		}
	}
}

// String implements [fmt.Stringer]
func (st Stack) String() string {
	var b strings.Builder

	for f := range st.All() {
		fmt.Fprintf(&b, "%s()\n", f.Function)
		fmt.Fprintf(&b, "\t%s:%d\n", f.File, f.Line)
	}

	return strings.TrimSpace(b.String())
}
