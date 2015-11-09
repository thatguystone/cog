package path

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestMain(m *testing.M) {
	check.Main(m)
}

func TestGenerateFromTypesBasic(t *testing.T) {
	c := check.New(t)

	if testing.Short() {
		c.Skip("skipping test in short mode.")
	}

	file := "file.go"
	c.FS.SWriteFile(file, fixtureBasic)

	err := GenerateFrom(c.FS.Path(file))
	c.MustNotError(err)

	s := c.FS.SReadFile(genFileName(file))

	c.Contains(s, `append(s.B, "static"...)`)
	c.Contains(s, "v.H.Marshal")
	c.Contains(s, "v.I.BoolInterfaced.Marshal")
	c.Contains(s, "s.EmitUint32(v.I.O)")
	c.Contains(s, "func (v *stuff) UnmarshalPath(s path.Decoder) path.Decoder {")

	// Exported fields shouldn't be around
	c.NotContains(s, "v.g")
}

func TestGenerateEndToEnd(t *testing.T) {
	c := check.New(t)

	if testing.Short() {
		c.Skip("skipping test in short mode.")
	}

	c.FS.SWriteFile("fixture.go", fixtureEndToEnd)
	c.FS.SWriteFile("integrate.go", fixtureIntegrate)
	c.FS.SWriteFile("integrate_test.go", fixtureEndToEndTest)

	err := GenerateFrom(c.FS.Path("integrate.go"))
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
