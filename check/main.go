package check

import (
	"flag"
	"os"
	"testing"

	"github.com/thatguystone/cog/cfs"
)

//gocovr:skip-file

// Main provides an alternative main for testing. This sets the testing
// environment up and also cleans it up on success; on failure, test files are
// left so that you can better inspect the failure.
//
// If you want to use this functionality, add the following somewhere in your
// tests:
//
//     func TestMain(m *testing.M) {
//     	check.Main(m)
//     }
func Main(m *testing.M) {
	flag.Parse()

	Setup()
	code := m.Run()
	if code == 0 {
		Cleanup()
	}

	os.Exit(code)
}

// Setup does basic pre-test checks to ensure the environment is ready to run.
// (It's just simple stuff, like make sure DataDir exists, and etc).
func Setup() {
	_, err := cfs.FindDir(DataDir)
	if err != nil {
		err = os.Mkdir(DataDir, 0750)
	}

	if err != nil {
		panic(err)
	}
}

// Cleanup removes any and all testing state
func Cleanup() {
	path, err := cfs.FindDir(DataDir)
	if err == nil {
		os.RemoveAll(path)
	}
}
