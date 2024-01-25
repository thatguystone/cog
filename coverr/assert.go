package coverr

import (
	"errors"
	"fmt"
)

// A Tester typically a [testing.T], [testing.B], or [testing.F]
type Tester interface {
	Helper()
	Fatal(args ...any)
}

// UntilNil is like [check.UntilNil], except it fails if a returned error isn't
// a [coverr.Err].
func UntilNil(t Tester, iters int, fn func(i int) error) bool {
	var err error

	for i := range iters {
		err = fn(i)
		if err == nil {
			return true
		}

		if !errors.Is(err, Err) {
			t.Helper()
			t.Fatal(
				fmt.Sprintf(
					"Func failed with non-coverr error: %v",
					err,
				),
			)
			return false
		}
	}

	t.Helper()
	t.Fatal(
		fmt.Sprintf(
			"Func didn't succeed after %d tries, last err: %v",
			iters,
			err,
		),
	)

	return false
}
