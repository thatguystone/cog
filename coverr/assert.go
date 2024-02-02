package coverr

import (
	"errors"
	"fmt"

	"github.com/thatguystone/cog/check"
)

func eventuallyNil(iters int, fn func(i int) error) (string, bool) {
	var err error

	for i := range iters {
		err = fn(i)
		if err == nil {
			return "", true
		}

		if !errors.Is(err, Err) {
			msg := fmt.Sprintf("Func failed with non-coverr error: %v", err)
			return msg, false
		}
	}

	msg := fmt.Sprintf(
		"Func didn't succeed after %d tries, last err: %v",
		iters,
		err,
	)
	return msg, false
}

// Like [check.EventuallyNil], except it also fails if a returned error isn't a
// [coverr.Err].
func EventuallyNil(t check.Error, iters int, fn func(i int) error) bool {
	if msg, ok := eventuallyNil(iters, fn); !ok {
		t.Helper()
		t.Error(msg)
		return false
	}

	return true
}

// Like [check.EventuallyNil], except it also fails if a returned error isn't a
// [coverr.Err].
func MustEventuallyNil(t check.Fatal, iters int, fn func(i int) error) {
	if msg, ok := eventuallyNil(iters, fn); !ok {
		t.Helper()
		t.Fatal(msg)
	}
}
