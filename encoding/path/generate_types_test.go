package path

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/iheartradio/cog/cfs"
	"github.com/iheartradio/cog/check"
)

func TestMain(m *testing.M) {
	check.Main(m)
}

func TestGenerateFromTypesBasic(t *testing.T) {
	c := check.New(t)

	if testing.Short() {
		c.Skip("skipping test in short mode.")
	}

	subPath, err := cfs.ImportPath(c.FS.Path("subpkg/subpkg.go"), false)
	c.MustNotError(err)

	otherPath, err := cfs.ImportPath(c.FS.Path("subother/other.go"), false)
	c.MustNotError(err)

	file := "file.go"
	c.FS.SWriteFile(file, fmt.Sprintf(fixtureBasic, subPath))

	c.FS.SWriteFile("subother/other.go", fixtureSubOther)
	c.FS.SWriteFile("subpkg/subpkg.go", fmt.Sprintf(fixtureSubpkg, otherPath))

	err = GenerateFrom(c.FS.Path(file))
	c.MustNotError(err)

	s := c.FS.SReadFile(genFileName(file))

	c.Contains(s, `append(s.B, "static"...)`)
	c.Contains(s, "v.H.Marshal")
	c.Contains(s, "v.I.BoolInterfaced.Marshal")
	c.Contains(s, "s.EmitUint32(uint32(v.I.O))")
	c.Contains(s, "func (v *stuff) UnmarshalPath(s path.Decoder) path.Decoder {")
	c.Contains(s, "v.L.A")
	c.Contains(s, "v.M.MarshalPath")
	c.Contains(s, "s = s.ExpectString((*string)(&v.SelectorExpr[i]))")
	c.Contains(s, "func (v basicRedef) MarshalPath")

	// Exported fields shouldn't be around
	c.NotContains(s, "v.g")
}

func TestGenerateEndToEnd(t *testing.T) {
	c := check.New(t)

	if testing.Short() {
		c.Skip("skipping test in short mode.")
	}

	subPath, err := cfs.ImportPath(c.FS.Path("subpkg/subpkg.go"), false)
	c.MustNotError(err)

	otherPath, err := cfs.ImportPath(c.FS.Path("subother/other.go"), false)
	c.MustNotError(err)

	c.FS.SWriteFile("fixture.go", fixtureEndToEnd)
	c.FS.SWriteFile("integrate.go",
		fmt.Sprintf(fixtureIntegrate, subPath))
	c.FS.SWriteFile("integrate_test.go", fixtureEndToEndTest)
	c.FS.SWriteFile("subother/other.go", fixtureSubOther)
	c.FS.SWriteFile("subpkg/subpkg.go", fmt.Sprintf(fixtureSubpkg, otherPath))

	err = GenerateFrom(c.FS.Path("integrate.go"))
	c.MustNotError(err)

	wd, err := os.Getwd()
	c.MustNotError(err)

	rel, err := filepath.Rel(wd, c.FS.Path(""))
	c.MustNotError(err)

	output, err := exec.Command("go", "test", "./"+rel).CombinedOutput()
	c.MustNotError(err, string(output))
}

func TestGenerateFromTypesErrors(t *testing.T) {
	c := check.New(t)

	err := GenerateFrom("blah blah blah")
	c.MustError(err)
}
