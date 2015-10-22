package cfs

import (
	"os"
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

func TestDirExists(t *testing.T) {
	t.Parallel()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	ex, err := DirExists(dir)
	if err != nil {
		panic(err)
	}

	if !ex {
		t.Fail()
	}

	ex, err = DirExists("/i/dont/exist")
	if err != nil {
		panic(err)
	}

	if ex {
		t.Fail()
	}
}

func TestTempFile(t *testing.T) {
	t.Parallel()

	f, err := TempFile("tmp-", ".aac")
	if err != nil {
		panic(err)
	}

	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	if !strings.Contains(f.Name(), "tmp-") {
		t.Fail()
	}

	if !strings.Contains(f.Name(), ".aac") {
		t.Fail()
	}
}
