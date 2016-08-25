package cort_test

import (
	"fmt"

	"github.com/iheartradio/cog/cort"
)

type IntSlice []int

func (p IntSlice) Len() int           { return len(p) }
func (p IntSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p IntSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p IntSlice) Move(i, j, a0, a1, b0, b1 int) {
	e := p[i]
	copy(p[a0:a1], p[b0:b1])
	p[j] = e
}

func Example_fix() {
	s := IntSlice{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	const i = 4

	for i := 1; i <= 5; i++ {
		s[i] = 10 * i
		fmt.Println("i = ", i)
		fmt.Println("Before:\t", s)
		cort.Fix(i, s)
		fmt.Println("After:\t", s)
		fmt.Println()
	}

	// Output:
	// i =  1
	// Before:	 [0 10 2 3 4 5 6 7 8 9]
	// After:	 [0 2 3 4 5 6 7 8 9 10]
	//
	// i =  2
	// Before:	 [0 2 20 4 5 6 7 8 9 10]
	// After:	 [0 2 4 5 6 7 8 9 10 20]
	//
	// i =  3
	// Before:	 [0 2 4 30 6 7 8 9 10 20]
	// After:	 [0 2 4 6 7 8 9 10 20 30]
	//
	// i =  4
	// Before:	 [0 2 4 6 40 8 9 10 20 30]
	// After:	 [0 2 4 6 8 9 10 20 30 40]
	//
	// i =  5
	// Before:	 [0 2 4 6 8 50 10 20 30 40]
	// After:	 [0 2 4 6 8 10 20 30 40 50]
}
