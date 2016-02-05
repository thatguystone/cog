package stack

import (
	"strings"
	"testing"
)

func TestClearTestCallerCoverage(t *testing.T) {
	t.Parallel()
	ClearTestCaller()
}

func testCaller(t *testing.T) {
	f := Caller(2)
	if strings.Contains(f, "caller_test.go") {
		t.Errorf("%s is wrong", f)
	}
}

func TestCaller(t *testing.T) {
	t.Parallel()

	f := Caller(0)
	if !strings.Contains(f, "caller_test.go") {
		t.Errorf("%s is wrong", f)
	}

	f = Caller(1)
	if strings.Contains(f, "caller_test.go") {
		t.Errorf("%s is wrong", f)
	}

	testCaller(t)
}

func TestCallerAbove(t *testing.T) {
	t.Parallel()

	d := CallerAbove(0, "testing")
	fl := Caller(d)
	if "???:1" == fl {
		t.Errorf("%d should resolve to something", d)
	}

	d = CallerAbove(0, "")
	fl = Caller(d)
	if "???:1" != fl {
		t.Errorf("%s should resolve to nothing", fl)
	}
}
