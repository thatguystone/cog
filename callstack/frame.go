package callstack

import (
	"runtime"

	"github.com/thatguystone/cog/assert"
)

// Frame wraps [runtime.Frame] with extra functionality
type Frame struct {
	runtime.Frame
}

// Self gets the Frame of the caller
func Self() Frame {
	return Caller(1)
}

// Caller gets the Frame of the caller after skipping the given number of frames
func Caller(skip int) Frame {
	var pcs [1]uintptr
	n := runtime.Callers(skip+2, pcs[:]) // runtime.Callers + Self()
	assert.True(n > 0)

	frames := runtime.CallersFrames(pcs[:])
	frame, _ := frames.Next()
	return Frame{frame}
}

// FuncName gets the non-qualified name of the function.
func (fr Frame) FuncName() string {
	_, funcName := fr.PkgAndFunc()
	return funcName
}

// PkgPath gets the name of the package the frame belongs to
func (fr Frame) PkgPath() string {
	pkgPath, _ := fr.PkgAndFunc()
	return pkgPath
}

func (fr Frame) PkgAndFunc() (pkgPath string, funcName string) {
	// Borrowed from [runtime.funcpkgpath]
	name := fr.Function
	i := len(name) - 1
	for ; i > 0; i-- {
		if name[i] == '/' {
			break
		}
	}
	for ; i < len(name); i++ {
		if name[i] == '.' {
			break
		}
	}
	return name[:i], name[i+1:]
}
