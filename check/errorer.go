package check

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Errorer is useful for mocking out things that return errors. It will return
// an error for every unique stack trace that it sees, but only on the first
// run. This allows you to run the same code many times in succession until it
// succeeds. By doing this, you can test that all your error pathways function
// correctly.
type Errorer struct {
	// If test functions should not be considered when comparing stack traces
	IgnoreTestFns bool

	// Only functions with these names should be considered for errors
	OnlyIn []string

	mtx    sync.Mutex
	stacks map[[sha256.Size]byte]struct{}
}

func (er *Errorer) fail() bool {
	hash := sha256.New()

	fail := false

	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Only worry about actual code paths, not where functions are called
		// from testing
		if er.IgnoreTestFns && strings.HasSuffix(file, "_test.go") {
			break
		}

		if len(er.OnlyIn) == 0 {
			fail = true
		} else if !fail {
			name := runtime.FuncForPC(pc).Name()

			for _, fn := range er.OnlyIn {
				fail = strings.Contains(name, fn)
				if fail {
					break
				}
			}
		}

		hash.Write([]byte(fmt.Sprintf("%d-%s:%d", pc, file, line)))
	}

	if !fail {
		return false
	}

	sum := [sha256.Size]byte{}
	copy(sum[:], hash.Sum(nil))

	er.mtx.Lock()
	defer er.mtx.Unlock()

	if er.stacks == nil {
		er.stacks = map[[sha256.Size]byte]struct{}{}
	}

	if _, ok := er.stacks[sum]; !ok {
		er.stacks[sum] = struct{}{}
		return true
	}

	return false
}

// Fail determines if the operation should fail with an error. This also marks
// the current stack as hit, so any future calls will return falser.
func (er *Errorer) Fail() bool {
	// Just bounce over to fail(): fail assumes a depth of 2, so can't just put
	// that logic here.
	return er.fail()
}

// Err returns a generic error if this should fail, or nil
func (er *Errorer) Err() (err error) {
	if er.fail() {
		err = errors.New("forced to fail by Errorer")
	}

	return
}

// UntilNil is a helper func used to exhaust error pathways when using an
// Errorer.
func UntilNil(timeout time.Duration, fn func() error) {
	var err error

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		err = fn()
		if err == nil {
			return
		}
	}

	panic(fmt.Errorf("func never succeeded, last err: %v", err))
}
