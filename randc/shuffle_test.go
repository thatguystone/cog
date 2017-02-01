package randc

import (
	"sort"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestShuffleBasic(t *testing.T) {
	c := check.New(t)

	orig := make(sort.IntSlice, 20)
	shuff := make(sort.IntSlice, 20)
	for i := range shuff {
		orig[i] = i
		shuff[i] = i
	}

	Shuffle(shuff)

	c.NotEqual(orig, shuff)
}
