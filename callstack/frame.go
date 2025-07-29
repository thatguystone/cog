package callstack

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// Raw program counter wrapper
type PC [1]uintptr

// Self gets the PC of the caller
func Self() PC {
	return Caller(1)
}

// Caller gets the PC of the caller after skipping the given number of frames
func Caller(skip int) (pc PC) {
	runtime.Callers(skip+2, pc[:]) // runtime.Callers + Self()
	return
}

func (pc PC) Frame() Frame {
	frames := runtime.CallersFrames(pc[:])
	frame, _ := frames.Next()
	return Frame{frame}
}

// Frame wraps [runtime.Frame] with extra functionality
type Frame struct {
	f runtime.Frame
}

// PC gets the raw program counter
func (frame Frame) PC() uintptr {
	return frame.f.PC
}

// Func gets the fully-qualified name of the function
func (frame Frame) Func() string {
	name := frame.f.Function
	if name == "" {
		return "???"
	}

	return name
}

// FuncName gets the non-qualified name of the function.
func (frame Frame) FuncName() string {
	_, funcName := frame.PkgAndFunc()
	return funcName
}

// PkgPath gets the name of the package the frame belongs to
func (frame Frame) PkgPath() string {
	pkgPath, _ := frame.PkgAndFunc()
	return pkgPath
}

// File gets the path and file name of this Frame
func (frame Frame) File() string {
	file := frame.f.File
	if file == "" {
		return "???"
	}

	return file
}

// FileName gets the file name of this Frame
func (frame Frame) FileName() string {
	file := frame.File()
	return filepath.Base(file)
}

// Line gets the line number of this Frame
func (frame Frame) Line() int {
	return frame.f.Line
}

// PkgAndFunc gets the name of the package this frame belongs to and the
// non-qualified name of the function in the package.
func (frame Frame) PkgAndFunc() (pkgPath string, funcName string) {
	name := frame.f.Function
	if name == "" {
		return "???", "???"
	}

	// Borrowed from [runtime.funcpkgpath]
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

func (frame Frame) append(b *strings.Builder) {
	fmt.Fprintf(b, "%s()\n", frame.Func())
	fmt.Fprintf(b, "\t%s:%d\n", frame.File(), frame.Line())
}

// String implements [fmt.Stringer]
func (frame Frame) String() string {
	var b strings.Builder
	frame.append(&b)
	return b.String()
}
