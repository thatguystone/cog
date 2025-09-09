package callstack

import (
	"iter"
	"runtime"
	"strings"
)

// A Stack is a stack trace
type Stack []uintptr

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
			return pcs[:n]
		}

		pcs = append(pcs, make([]uintptr, len(pcs))...)
	}
}

// Frames returns an iterator over every [Frame] in the stack.
func (st Stack) Frames() iter.Seq[Frame] {
	return func(yield func(Frame) bool) {
		frames := runtime.CallersFrames(st)
		for {
			frame, more := frames.Next()
			if frame != (runtime.Frame{}) {
				if !yield(Frame{f: frame}) {
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
	for frame := range st.Frames() {
		frame.append(&b)
	}

	return strings.TrimSpace(b.String())
}
