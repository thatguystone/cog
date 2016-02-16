package cfs_test

import (
	"testing"

	"github.com/thatguystone/cog/cfs"
	"github.com/thatguystone/cog/check"
)

func TestCreateParents(t *testing.T) {
	c := check.New(t)

	parents := c.FS.Path("really/long/path/with/parents")
	err := cfs.CreateParents(parents + "/file")
	c.MustNotError(err)

	exists, err := cfs.DirExists(parents)
	c.MustNotError(err)
	c.True(exists)
}
