package cfs_test

import (
	"testing"

	"github.com/thatguystone/cog/cfs"
	"github.com/thatguystone/cog/check"
)

func TestCreateParents(t *testing.T) {
	c := check.New(t)

	fs, clean := c.FS()
	defer clean()

	parents := fs.Path("really/long/path/with/parents")
	err := cfs.CreateParents(parents + "/file")
	c.Must.Nil(err)

	exists, err := cfs.DirExists(parents)
	c.Must.Nil(err)
	c.True(exists)
}
