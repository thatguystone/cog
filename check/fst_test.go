package check

import "testing"

func TestFSBasic(t *testing.T) {
	c := New(t)

	cs := "file contents"

	c.FS.SWriteFile("test", cs)
	got := c.FS.SReadFile("test")

	c.Equal(cs, got)
}
