package stats

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestJoin(t *testing.T) {
	c := check.New(t)

	c.Equal("a.b.c", Join("......a....b....c......"))
}
