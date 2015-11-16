// Package cort implements extra sort helpers.
//
// "cort" is pronounced "sort".
package cort

import (
	"math/rand"
	"sort"
	"time"
)

// A Mover moves a single value around inside of itself, given a subslice of
// itself to shift.
type Mover interface {
	sort.Interface
	Move(i, j, a0, a1, b0, b1 int)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Fix updates an already-sorted slice after an element has been updated. Rather
// than calling sort.Sort again (which is painfully slow for already-sorted
// slices), this very quickly finds a new home for the changed element.
func Fix(i int, s Mover) {
	n := s.Len() - 1
	to := sort.Search(n, func(j int) bool {
		if j >= i {
			return s.Less(i, j+1)
		}

		return s.Less(i, j)
	})

	var a0, a1, b0, b1 int
	if to > i {
		a0 = i
		a1 = to
		b0 = i + 1
		b1 = to + 1
	} else {
		a0 = to + 1
		a1 = i + 1
		b0 = to
		b1 = i
	}

	if to == i {
		return
	}

	s.Move(i, to, a0, a1, b0, b1)
}
