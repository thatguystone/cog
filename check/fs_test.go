package check

import "testing"

func TestFSBasic(t *testing.T) {
	c := New(t)

	fs, clean := c.FS()
	defer clean()

	cs := "file contents"
	fs.SWriteFile("test", cs)
	got := fs.SReadFile("test")

	c.Equal(cs, got)
}

func TestFSRefs(t *testing.T) {
	c := New(t)

	fs, clean := c.FS()
	fs.SWriteFile("dir/file", "")

	fs2, clean2 := fs.ref()
	fs2.SWriteFile("dir/file2", "")

	fs.FileExists("dir/file")
	fs.FileExists("dir/file2")

	clean2()

	fs.FileExists("dir/file")
	fs.FileExists("dir/file2")

	clean()

	// Everything unrefd: new FS() should get a clean dir
	fs, clean = c.FS()
	defer clean()
	fs.NotDirExists("dir")
}

func TestNewFS(t *testing.T) {
	c := New(t)

	fs0, clean := c.NewFS()
	defer clean()

	fs1, clean := c.NewFS()
	defer clean()

	fs0.SWriteFile("test", "")
	fs1.NotFileExists("test")
}

func TestCleanDir(t *testing.T) {
	c := New(t)

	fs, clean := c.FS()
	defer clean()

	c.Panics(func() {
		fs.cleanDir("/this/doesnt/exist")
	})
}

func TestCleanDataDir(t *testing.T) {
	c := New(t)

	fs, clean := c.FS()
	defer clean()

	fs.SWriteFile("dir/file", "")
	fs.CleanDataDir()
	fs.NotFileExists("dir/file")
}

func TestFSContentsEqual(t *testing.T) {
	c := New(t)

	fs, clean := c.FS()
	defer clean()

	cs := "file contents"
	fs.SWriteFile("test", cs)

	fs.ContentsEqual("test", []byte(cs))
	fs.SContentsEqual("test", cs)
}

func TestFSExists(t *testing.T) {
	c := New(t)

	fs, clean := c.FS()
	defer clean()

	fs.SWriteFile("dir/file", "")

	c.True(fs.FileExists("dir/file"))
	c.True(fs.DirExists("dir"))

	c.True(fs.NotFileExists("file"))
	c.True(fs.NotDirExists("dir2"))
}

func TestDumpTree(t *testing.T) {
	c := New(t)

	fs, clean := c.FS()
	defer clean()

	fs.SWriteFile("dir/file", "")
	fs.SWriteFile("dir2/file", "")

	fs.DumpTree("")
}
