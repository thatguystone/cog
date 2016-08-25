package clog_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/iheartradio/cog/clog"
)

func Example_file() {
	path := "example_file.log"
	cfg := clog.Config{
		File: path,
	}

	l, err := clog.New(cfg)
	if err != nil {
		panic(err)
	}

	defer os.Remove(path)

	lg := l.Get("example")

	lg.Debug("Bug bug bug bug")
	lg.Info("This is a very informative message")
	lg.WithKV("variable", 1234).Warn("I must warn you, numbers are scary")

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	log := string(contents)
	fmt.Println("contains debug", strings.Contains(log, "Bug bug bug bug"))
	fmt.Println("contains info", strings.Contains(log, "informative message"))
	fmt.Println("contains warn", strings.Contains(log, "I must warn you"))

	// Output:
	// contains debug false
	// contains info true
	// contains warn true
}
