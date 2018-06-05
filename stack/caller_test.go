package stack

import (
	"strings"
	"testing"
)

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
