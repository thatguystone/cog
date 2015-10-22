package cfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindDirInParents(t *testing.T) {
	t.Parallel()

	_, err := FindDirInParents("idontexist")
	if err == nil {
		t.Fail()
	}

	_, err = FindDirInParents("cog")
	if err != nil {
		t.Fail()
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	fs, err := filepath.Glob("*")
	if len(fs) == 0 || err != nil {
		t.Errorf("could not find files: len=%d, err=%v", len(fs), err)
	}

	ex, err := FileExists(fs[0])
	if !ex || err != nil {
		t.Errorf("%s should exist: ex=%t, err=%v", fs[0], ex, err)
	}

	ex, err = FileExists("/i/dont/exist")
	if ex || err != nil {
		t.Errorf("should not exist: ex=%t, err=%v", ex, err)
	}
}

func TestDirExists(t *testing.T) {
	t.Parallel()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd: %v", err)
	}

	ex, err := DirExists(dir)
	if !ex || err != nil {
		t.Errorf("%s should exist: ex=%t err=%v", dir, ex, err)
	}

	ex, err = DirExists("/i/dont/exist/")
	if ex || err != nil {
		t.Errorf("%s should not exist: ex=%t err=%v", dir, ex, err)
	}
}

func TestTempFile(t *testing.T) {
	t.Parallel()

	f, err := TempFile("tmp-", "aac")
	if err != nil {
		t.Fatalf("failed to create TempFile: %v", err)
	}

	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	if !strings.Contains(f.Name(), "tmp-") {
		t.Errorf("%s doesn't contain %s", f.Name(), "tmp-")
	}

	if !strings.Contains(f.Name(), ".aac") {
		t.Errorf("%s doesn't contain %s", f.Name(), ".aac")
	}
}
