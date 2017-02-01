package randc

import (
	"math/rand"
	"sort"
)

// Shuffle implements a Fisher-Yates shuffle for slices that implement
// sort.Interface.
func Shuffle(data sort.Interface) {
	n := data.Len()

	for i := 0; i < n; i++ {
		j := rand.Intn(i + 1)
		data.Swap(i, j)
	}
}
