// Package check provides dead-simple assertions and utilities for testing.
package check

import (
	"reflect"
	"testing"
)

// A T wraps a [testing.T]
type T struct {
	*testing.T
	assert
	Must assert
}

// New creates a new T and marks this test as parallel
func NewT(t *testing.T) *T {
	t.Parallel()

	return &T{
		T:      t,
		assert: newAssert(t.Helper, t.Error),
		Must:   newAssert(t.Helper, t.Fatal),
	}
}

// Run is the equivalent of [testing.T.Run]
func (t *T) Run(name string, fn func(*T)) bool {
	return t.T.Run(name, func(t *testing.T) {
		fn(NewT(t))
	})
}

// A B wraps a [testing.B]
type B struct {
	*testing.B
	assert
	Must assert
}

// NewB creates a new B
func NewB(b *testing.B) *B {
	return &B{
		B:      b,
		assert: newAssert(b.Helper, b.Error),
		Must:   newAssert(b.Helper, b.Fatal),
	}
}

// Run is the equivalent of [testing.B.Run]
func (b *B) Run(name string, fn func(*B)) bool {
	return b.B.Run(name, func(b *testing.B) {
		fn(NewB(b))
	})
}

// A F wraps a [testing.F]
type F struct {
	*testing.F
	assert
	Must assert
}

// NewF creates a new F
func NewF(f *testing.F) *F {
	return &F{
		F:      f,
		assert: newAssert(f.Helper, f.Error),
		Must:   newAssert(f.Helper, f.Fatal),
	}
}

var (
	tPtr  = reflect.TypeOf((*testing.T)(nil))
	ctPtr = reflect.TypeOf((*T)(nil))
)

// Run is the equivalent of [testing.F.Fuzz]
func (f *F) Fuzz(fn any) {
	fv := reflect.ValueOf(fn)
	ft := fv.Type()

	if ft.Kind() != reflect.Func {
		panic("check: F.Fuzz must receive a function")
	}

	if ft.NumIn() < 2 || ft.In(0) != ctPtr {
		panic("check: fuzz fn must be in form func(*check.T, ...)")
	}

	if ft.NumOut() != 0 {
		panic("check: fuzz fn must not return a value")
	}

	args := make([]reflect.Type, ft.NumIn())
	args[0] = tPtr
	for i := 1; i < len(args); i++ {
		args[i] = ft.In(i)
	}

	wfv := reflect.MakeFunc(
		reflect.FuncOf(args, nil, false),
		func(args []reflect.Value) (results []reflect.Value) {
			ct := NewT(args[0].Interface().(*testing.T))
			args[0] = reflect.ValueOf(ct)

			return fv.Call(args)
		},
	)

	f.F.Fuzz(wfv.Interface())
}
