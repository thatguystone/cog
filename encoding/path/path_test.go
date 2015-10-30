package path

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"testing"

	"github.com/thatguystone/cog/check"
)

type Empty struct{}

type Pie struct {
	Start Static `path:"pie"`
	A     string
	B     uint8
	C     string
	D     uint8
	E     string
	F     uint8
	End   Static `path:"end"`
	g     string
	h     uint8
}

type Everything struct {
	A int8
	B int16
	C int32
	D int64
	E uint8
	F uint16
	G uint32
	H uint64
	I float32
	J float64
	K complex64
	L complex128
	M bool
}

type MinMax struct {
	I int8
	U uint8
}

type NestingI struct {
	A int8
	B int8
}

type NestingU struct {
	A uint8
	B uint8
}

type Nesting struct {
	NestingI
	NestingU
}

func trampoline(c *check.C, in interface{}, out interface{}, path []byte) {
	p, err := Marshal(in)
	c.MustNotError(err)

	if path != nil {
		c.Equal(path, p)
	}

	err = Unmarshal(p, out)
	c.MustNotError(err)
	c.Equal(in, out)
}

func TestEmpty(t *testing.T) {
	c := check.New(t)

	e := Empty{}
	e2 := Empty{}

	trampoline(c, &e, &e2, []byte("/"))
}

func TestNil(t *testing.T) {
	c := check.New(t)

	_, err := Marshal(nil)
	c.Error(err)

	e := Everything{}
	err = Unmarshal(nil, &e)
	c.Error(err)
}

func TestNonStruct(t *testing.T) {
	c := check.New(t)

	i := int(0)
	_, err := Marshal(i)
	c.Error(err)
}

func TestNonStructPtr(t *testing.T) {
	c := check.New(t)

	var i *int
	_, err := Marshal(i)
	c.Error(err)
}

func TestPieTrampoline(t *testing.T) {
	c := check.New(t)

	pie := Pie{
		A: "apple",
		B: 1,
		C: "pumpkin",
		D: 2,
		E: "berry",
		F: 3,
		g: "nope",
		h: 123,
	}

	p, err := Marshal(pie)
	c.MustNotError(err)
	c.Equal("/pie/apple/\x01/pumpkin/\x02/berry/\x03/end/", string(p))

	p2 := Pie{}
	err = Unmarshal(p, &p2)
	c.MustNotError(err)
	c.Equal(pie.A, p2.A)
	c.Equal(pie.B, p2.B)
	c.Equal(pie.C, p2.C)
	c.Equal(pie.D, p2.D)
	c.Equal(pie.E, p2.E)
	c.Equal(pie.F, p2.F)
	c.Equal("", p2.g)
	c.Equal(0, p2.h)
}

func TestEverything(t *testing.T) {
	c := check.New(t)

	e := Everything{
		A: math.MinInt8,
		B: math.MinInt16,
		C: math.MinInt32,
		D: math.MinInt64,
		E: math.MaxUint8,
		F: math.MaxUint16,
		G: math.MaxUint32,
		H: math.MaxUint64,
		I: float32(1.2),
		J: float64(1123131.298488474),
		K: 1 + 0i,
		L: 1 + 0i,
		M: true,
	}

	e2 := Everything{}

	trampoline(c, &e, &e2, nil)
}

func TestSort(t *testing.T) {
	c := check.New(t)

	vs := []struct {
		A uint8
		B string
	}{
		{
			A: 2,
			B: "b",
		},
		{
			A: 2,
			B: "a",
		},
		{
			A: 1,
			B: "f",
		},
		{
			A: 1,
			B: "e",
		},
	}

	bs := []string{}

	for _, v := range vs {
		path, err := Marshal(v)
		c.MustNotError(err)

		bs = append(bs, string(path))
	}

	sort.Sort(sort.StringSlice(bs))

	ss := []string{
		"/\x01/e/",
		"/\x01/f/",
		"/\x02/a/",
		"/\x02/b/",
	}
	c.Equal(ss, bs)
}

func TestFuzz(t *testing.T) {
	c := check.New(t)

	wg := sync.WaitGroup{}
	for i := 0; i < 2048; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			e := Everything{
				A: int8(rand.Int63()),
				B: int16(rand.Int63()),
				C: int32(rand.Int63()),
				D: rand.Int63(),
				E: uint8(rand.Uint32()),
				F: uint16(rand.Uint32()),
				G: rand.Uint32(),
				H: uint64(rand.Uint32())<<32 | uint64(rand.Uint32()),
				I: rand.Float32(),
				J: rand.Float64(),
				K: complex(rand.Float32(), rand.Float32()),
				L: complex(rand.Float64(), rand.Float64()),
				M: rand.Int()%2 == 0,
			}

			e2 := Everything{}

			trampoline(c, &e, &e2, nil)
		}()
	}

	wg.Wait()
}

func TestNesting(t *testing.T) {
	c := check.New(t)

	n := Nesting{
		NestingI: NestingI{
			A: 1,
			B: 2,
		},
		NestingU: NestingU{
			A: 3,
			B: 4,
		},
	}

	n2 := Nesting{}

	trampoline(c, &n, &n2, []byte("/\x01/\x02/\x03/\x04/"))
}

func TestMinMax(t *testing.T) {
	c := check.New(t)

	i8 := int8(math.MinInt8)
	u8 := uint8(0)

	wg := sync.WaitGroup{}
	for i := 0; i <= math.MaxUint8; i++ {
		wg.Add(1)
		go func(i8 int8, u8 uint8) {
			wg.Done()
			mm := MinMax{
				I: i8,
				U: u8,
			}

			mm2 := MinMax{}

			trampoline(c, &mm, &mm2, nil)
		}(i8, u8)

		i8++
		u8++
	}

	wg.Wait()
}

func TestUnsupportedType(t *testing.T) {
	c := check.New(t)

	v := struct {
		A chan int
	}{}

	_, err := Marshal(v)
	c.Error(err)

	err = Unmarshal([]byte("/asd"), &v)
	c.Error(err)
}

func TestInvalidString(t *testing.T) {
	c := check.New(t)

	v := struct {
		A string
	}{
		A: "string/with/slashes",
	}

	_, err := Marshal(v)
	c.Error(err)
}

func TestTruncatedInt(t *testing.T) {
	c := check.New(t)

	v := struct{ A uint64 }{}

	err := Unmarshal([]byte("/a"), &v)
	c.Error(err)
}

func TestUnmarshalWithoutPointer(t *testing.T) {
	c := check.New(t)
	v := struct{ A int }{}

	err := Unmarshal(nil, v)
	c.Error(err)
}

func TestWrongTag(t *testing.T) {
	c := check.New(t)

	in := struct {
		A Static `path:"in"`
	}{}
	out := struct {
		A Static `path:"out"`
	}{}

	b, err := Marshal(in)
	c.MustNotError(err)

	err = Unmarshal(b, &out)
	c.Error(err)
}

func BenchmarkMarshal(b *testing.B) {
	pie := Pie{
		A: "apple",
		B: 1,
		C: "apple",
		D: 1,
		E: "apple",
		F: 1,
	}

	for i := 0; i < b.N; i++ {
		Marshal(pie)
	}
}

func Example_usage() {
	v := struct {
		Type         string
		Index        uint32
		Desirability uint64
	}{
		Type:         "apple",
		Index:        123,
		Desirability: 99828,
	}

	// This path isn't human-readable, so printing it anywhere it pointless
	path, err := Marshal(v)
	if err != nil {
		panic(err)
	}

	// Reset everything, just in case
	v.Type = ""
	v.Index = 0
	v.Desirability = 0

	err = Unmarshal(path, &v)
	if err != nil {
		panic(err)
	}

	fmt.Println("Type:", v.Type)
	fmt.Println("Index:", v.Index)
	fmt.Println("Desirability:", v.Desirability)

	// Output:
	// Type: apple
	// Index: 123
	// Desirability: 99828
}

func ExampleStatic() {
	v := struct {
		S     Static `path:"pie"` // Prefix this path with "/pie/"
		Type  string
		Taste string
	}{
		Type:  "apple",
		Taste: "fantastic",
	}

	path, err := Marshal(v)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(path))

	// Output:
	// /pie/apple/fantastic/
}
