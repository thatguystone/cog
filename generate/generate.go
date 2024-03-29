package generate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thatguystone/cog/assert"
	"golang.org/x/tools/imports"
)

type Buffer struct {
	bytes.Buffer
	dstPath string
}

func New() *Buffer {
	const (
		cmdSuffix = ".cmd.go"
		outSuffix = ".out.go"
	)

	srcPath := os.Getenv("GOFILE")
	if !strings.HasSuffix(srcPath, cmdSuffix) {
		panic(
			fmt.Errorf(
				`generate cmd %q must have a name in the form of "*%s"`,
				srcPath,
				cmdSuffix,
			),
		)
	}

	b := &Buffer{
		dstPath: strings.TrimSuffix(srcPath, cmdSuffix) + outSuffix,
	}

	fmt.Fprintf(b, "// Code generated by `go generate %s`. DO NOT EDIT.\n", srcPath)
	fmt.Fprintf(b, "\n")
	fmt.Fprintf(b, "package %s\n", getPkgName())
	fmt.Fprintf(b, "\n")

	return b
}

func (b *Buffer) WriteFile() {
	out, err := imports.Process(b.dstPath, b.Bytes(), nil)
	assert.Nil(err)

	err = os.WriteFile(b.dstPath, out, 0640)
	assert.Nil(err)
}

func getPkgName() string {
	buf := new(bytes.Buffer)

	cmd := exec.Command("go", "list", "-f={{ .Name }}")
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	assert.Nil(err)

	return strings.TrimSpace(buf.String())
}
