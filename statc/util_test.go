package statc

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestJoinPath(t *testing.T) {
	c := check.New(t)
	c.Equal("multi.parent.dir.file_jpg", JoinPath("multi.parent", "dir", "file.jpg"))
}

func TestCleanPath(t *testing.T) {
	c := check.New(t)
	c.Equal("a.b.c", CleanPath("......a....b....c......"))
}
