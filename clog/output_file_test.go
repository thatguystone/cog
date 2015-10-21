package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestOutputFileNew(t *testing.T) {
	c := check.New(t)

	fmttrs := []string{"logfmt", "human", "json"}
	for _, fmttr := range fmttrs {
		_, err := newFileOutputter(
			ConfigOutputArgs{
				"format": fmttr,
				"path":   c.FS.Path(fmttr),
			},
			nil)
		c.NotError(err, "%s failed", fmttr)
	}

	_, err := newFileOutputter(
		ConfigOutputArgs{
			"format": "gerp gorp",
			"path":   c.FS.Path("gerp gorp"),
		},
		nil)
	c.Error(err)
}

func TestOutputFileRotate(t *testing.T) {
	// c := check.New(t)

	// l, err := newFileOutputter(
	// 	ConfigOutputArgs{
	// 		"format": "human",
	// 		"path":   c.FS.Path("log"),
	// 	},
	// 	nil)
	// c.MustNotError(err)
}
