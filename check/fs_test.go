package check

import "testing"

func TestFSBasic(t *testing.T) {
	c := New(t)

	cs := "file contents"

	c.FS.SWriteFile("test", cs)
	got := c.FS.SReadFile("test")

	c.Equal(cs, got)
}

func TestFSContentsEqual(t *testing.T) {
	c := New(t)

	cs := "file contents"
	c.FS.SWriteFile("test", cs)

	c.FS.ContentsEqual("test", []byte(cs))
	c.FS.SContentsEqual("test", cs)
}

func TestFSExists(t *testing.T) {
	c := New(t)

	c.FS.SWriteFile("dir/file", "")

	c.FS.FileExists("dir/file")
	c.FS.DirExists("dir")

	c.FS.FileNotExists("file")
	c.FS.DirNotExists("dir2")
}

func TestDumpTree(t *testing.T) {
	c := New(t)

	c.FS.SWriteFile("dir/file", "")
	c.FS.SWriteFile("dir2/file", "")

	c.FS.DumpTree("")
}
