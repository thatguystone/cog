package check

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func testDump(v any) string {
	return dump(v, 0)
}

func TestDumpZeros(t *testing.T) {
	var v struct {
		Bool    bool
		Int     int
		UInt    uint
		Float   float32
		Complex complex64
		Slice   []int
		Array   [3]int
		Map     map[string]string
		Ptr     *any
		Iface   error
		Chan    chan struct{}
		Func    func()
		Uintptr uintptr
		Struct  struct{}
	}

	_ = testDump(v)

	Equal(t, testDump(nil), "nil")
}

func TestDumpNamedTypes(t *testing.T) {
	type (
		namedBool   bool
		namedString string
		namedPtr    *string
	)

	var (
		b    = true
		nb   = namedBool(b)
		s    = "str"
		ns   = namedString(s)
		nptr = namedPtr(&s)
	)

	Equal(t, testDump(b), fmt.Sprintf("%v", b))
	Equal(t, testDump(nb), fmt.Sprintf("%T(%v)", nb, b))
	Equal(t, testDump(s), fmt.Sprintf("%q", s))
	Equal(t, testDump(ns), fmt.Sprintf("%T(%q)", ns, s))
	Equal(t, testDump(&nptr), fmt.Sprintf("&%T(&%q)", nptr, s))
}

func TestDumpNumbers(t *testing.T) {
	Equal(t, testDump(1), "int(1)")
	Equal(t, testDump(uint(1)), "uint(1)")
	Equal(t, testDump(1.0), "float64(1.0)")
	Equal(t, testDump(-1.0), "float64(-1.0)")
	Equal(t, testDump(math.NaN()), "float64(NaN)")
	Equal(t, testDump(1+3i), "complex128(1 + 3i)")
	Equal(t, testDump(1-3i), "complex128(1 - 3i)")
}

func TestDumpSlices(t *testing.T) {
	Equal(t, testDump([]int(nil)), "[]int(nil)")
	Equal(t, testDump([]int{}), "[]int{}")
}

func TestDumpBytes(t *testing.T) {
	Equal(
		t,
		testDump([]byte{}),
		"[]uint8{}",
	)
	Equal(
		t,
		testDump([]byte{1, 2, 3, 4, 5, 6, 7, 8}),
		"[]uint8{\n"+
			dumpIndent+"0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,\n"+
			"}",
	)
	Equal(
		t,
		testDump([]byte{0, 1, 2, 3}),
		"[]uint8{\n"+
			dumpIndent+"0x00, 0x01, 0x02, 0x03,\n"+
			"}",
	)
	Equal(
		t,
		testDump([]byte{
			0, 1, 2, 3, 4, 5, 6, 7,
			8, 9, 10, 11, 12, 13, 14, 15,
			16, 17, 18, 19, 20,
		}),
		"[]uint8{\n"+
			dumpIndent+"0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,\n"+
			dumpIndent+"0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,\n"+
			dumpIndent+"0x10, 0x11, 0x12, 0x13, 0x14,\n"+
			"}",
	)
}

func TestDumpMaps(t *testing.T) {
	var (
		mnil   = map[int]int(nil)
		mempty = map[int]int{}
	)

	Equal(t, testDump(mnil), "map[int]int(nil)")
	Equal(t, testDump(mempty), "map[int]int{}")
}

type testStringer string

func (s testStringer) String() string {
	return string(s)
}

type testError string

func (s testError) Error() string {
	return string(s)
}

type testPanics string

func (s testPanics) String() string {
	panic(string(s))
}

func TestDumpAnnotations(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		Equal(
			t,
			testDump(testStringer("henlo")),
			`/* "henlo" */check.testStringer("henlo")`,
		)
		Equal(
			t,
			testDump(testError("henlo")),
			`/* "henlo" */check.testError("henlo")`,
		)
	})

	t.Run("Panics", func(t *testing.T) {
		Equal(
			t,
			testDump(testPanics("test panic")),
			`/* (PANIC="test panic") */check.testPanics("test panic")`,
		)
	})
}

func TestDumpCircular(t *testing.T) {
	t.Run("Pointer", func(t *testing.T) {
		type circular struct {
			A *circular
			B *circular
			c *circular
		}

		var (
			p0 = new(circular)
			p1 = new(circular)
			p2 = new(circular)
		)

		p0.A = p0
		p0.B = p1
		p0.c = p2

		p1.A = p0
		p1.B = p2
		p1.c = p1

		p2.A = p1
		p2.B = p0
		p2.c = p2

		testDump(p0)
		testDump(p1)
		testDump(p2)
	})

	t.Run("Slice", func(t *testing.T) {
		t.Run("0", func(t *testing.T) {
			s := []any{}
			s = append(s, nil)
			s[0] = s
			testDump(s)
		})

		t.Run("1", func(t *testing.T) {
			s := make([]any, 1)
			e0 := &s[0]
			e1 := &e0
			s[0] = &e1
			testDump(s)
		})

		t.Run("2", func(t *testing.T) {
			var (
				x  any
				x0 = &x
				x1 = &x0
				x2 = &x1
				x3 = &x2
			)

			x = x1
			testDump([]any{x3, x2})
		})
	})

	t.Run("Map", func(t *testing.T) {
		m := make(map[any]any)
		m[1] = m
		m[&m] = 1
		testDump(m)
	})

	t.Run("Interface", func(t *testing.T) {
		s := make([]any, 1)
		s[0] = &s[0]
		testDump(s)
	})
}

func TestDumpGoString(t *testing.T) {
	Equal(t, testDump("plain"), `"plain"`)
	Equal(t, testDump(`"quotes"`), "`\"quotes\"`")
}

func TestFmtBase10(t *testing.T) {
	fmtInt64 := func(v int64) string {
		buf := make([]byte, 0, maxBase10Len)
		buf = strconv.AppendInt(buf, v, 10)
		return string(fmtBase10(buf))
	}
	fmtUint64 := func(v uint64) string {
		buf := make([]byte, 0, maxBase10Len)
		buf = strconv.AppendUint(buf, v, 10)
		return string(fmtBase10(buf))
	}

	Equal(t, fmtInt64(0), "0")
	Equal(t, fmtInt64(1), "1")
	Equal(t, fmtInt64(12), "12")
	Equal(t, fmtInt64(123), "123")
	Equal(t, fmtInt64(1_234), "1_234")
	Equal(t, fmtInt64(12_345), "12_345")
	Equal(t, fmtInt64(123_456), "123_456")
	Equal(t, fmtInt64(1_234_567), "1_234_567")

	Equal(t, fmtInt64(-1), "-1")
	Equal(t, fmtInt64(-12), "-12")
	Equal(t, fmtInt64(-123), "-123")
	Equal(t, fmtInt64(-1_234), "-1_234")
	Equal(t, fmtInt64(-12_345), "-12_345")
	Equal(t, fmtInt64(-123_456), "-123_456")
	Equal(t, fmtInt64(-1_234_567), "-1_234_567")

	Equal(t, fmtInt64(math.MaxInt64), "9_223_372_036_854_775_807")
	Equal(t, fmtInt64(math.MinInt64), "-9_223_372_036_854_775_808")
	Equal(t, fmtUint64(math.MaxUint64), "18_446_744_073_709_551_615")
}
