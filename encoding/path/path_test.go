package path

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/iheartradio/cog/check"
)

type Empty struct{}

type Pie struct {
	_ Static `path:"pie"`
	A string
	B uint8
	C string
	D uint8
	E string
	F uint8
	_ Static `path:"end"`
	g string
	h uint8
}

type PieWithInterface struct {
	Pie
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

type EverythingPtr struct {
	A *int8
	B *int16
	C *int32
	D *int64
	E *uint8
	F *uint16
	G *uint32
	H *uint64
	I *float32
	J *float64
	K *complex64
	L *complex128
	M *bool
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

type b bool
type i8 int8
type i16 int16
type i32 int32
type i64 int64
type u8 uint8
type u16 uint16
type u32 uint32
type u64 uint64
type f32 float32
type f64 float64
type c64 complex64
type c128 complex128

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (p PieWithInterface) MarshalPath(e Encoder) Encoder {
	e.B = append(e.B, "pie"...)
	e = e.EmitSep()
	e = e.EmitString(p.A)
	e = e.EmitUint8(p.B)
	e = e.EmitString(p.C)
	e = e.EmitUint8(p.D)
	e = e.EmitString(p.E)
	e = e.EmitUint8(p.F)
	e.B = append(e.B, "end"...)
	e = e.EmitSep()

	return e
}

var (
	pieWithInterfaceTag0 = []byte("pie")
	pieWithInterfaceTag1 = []byte("end")
)

func (p *PieWithInterface) UnmarshalPath(d Decoder) Decoder {
	d = d.ExpectTagBytes(pieWithInterfaceTag0)

	if d.Err == nil {
		d = d.ExpectString(&p.A)
	}

	if d.Err == nil {
		d = d.ExpectUint8(&p.B)
	}

	if d.Err == nil {
		d = d.ExpectString(&p.C)
	}

	if d.Err == nil {
		d = d.ExpectUint8(&p.D)
	}

	if d.Err == nil {
		d = d.ExpectString(&p.E)
	}

	if d.Err == nil {
		d = d.ExpectUint8(&p.F)
	}

	if d.Err == nil {
		d = d.ExpectTagBytes(pieWithInterfaceTag1)
	}

	return d
}

func trampoline(c *check.C, in interface{}, out interface{}, path []byte) {
	p, err := Marshal(in, nil)
	c.MustNotError(err)

	if path != nil {
		c.Equal(path, p)
	}

	_, err = Unmarshal(p, out)
	c.MustNotError(err)

	// Follow the ptr so that the Equal check works
	indirectOut := reflect.Indirect(reflect.ValueOf(out)).Interface()
	c.Equal(in, indirectOut)
}

func TestEmpty(t *testing.T) {
	c := check.New(t)

	e := Empty{}
	e2 := Empty{}

	trampoline(c, e, &e2, []byte("/"))
}

func TestNils(t *testing.T) {
	c := check.New(t)

	_, err := Marshal(nil, nil)
	c.Error(err)

	_, err = Marshal(EverythingPtr{}, nil)
	c.Error(err)

	e := Everything{}
	_, err = Unmarshal(nil, &e)
	c.Error(err)

	_, err = Unmarshal([]byte("/test/"), nil)
	c.Error(err)
}

func TestNonNilPtrs(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		new(bool),
		new(int8),
		new(int16),
		new(int32),
		new(int64),
		new(uint8),
		new(uint16),
		new(uint32),
		new(uint64),
		new(float32),
		new(float64),
		new(complex64),
		new(complex128),
		new(string),
	}

	for _, t := range tests {
		_, err := Marshal(t, nil)
		c.Error(err)
	}
}

func TestMarshalerInterface(t *testing.T) {
	c := check.New(t)

	p := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}
	p2 := PieWithInterface{}

	trampoline(c, p, &p2, nil)
}

func TestStateMusts(t *testing.T) {
	check.New(t)

	p := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}
	p2 := PieWithInterface{}

	b := p.MarshalPath(DefSep.NewEncoder(nil)).Must()
	p2.UnmarshalPath(DefSep.NewDecoder(b)).Must()
}

func TestNonStruct(t *testing.T) {
	c := check.New(t)

	i := int32(1)
	i2 := int32(0)

	trampoline(c, i, &i2, []byte("/\x00\x00\x00\x01/"))
}

func TestBytes(t *testing.T) {
	c := check.New(t)

	i := []byte("test")
	var i2 []byte
	out := []byte("/test/")

	trampoline(c, i, &i2, out)

	p, err := Marshal(i, nil)
	c.MustNotError(err)
	c.Equal(out, p)
	_, err = Unmarshal(p, &i2)
	c.MustNotError(err)
	c.Equal(i, i2)
}

func TestNonStructPtr(t *testing.T) {
	c := check.New(t)

	var i *int
	_, err := Marshal(i, nil)
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

	p, err := Marshal(pie, nil)
	c.MustNotError(err)
	c.Equal("/pie/apple/\x01/pumpkin/\x02/berry/\x03/end/", string(p))

	p2 := Pie{}
	_, err = Unmarshal(p, &p2)
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

	trampoline(c, e, &e2, nil)
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
		path, err := Marshal(v, nil)
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

			trampoline(c, e, &e2, nil)
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

	trampoline(c, n, &n2, []byte("/\x01/\x02/\x03/\x04/"))
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

			trampoline(c, mm, &mm2, nil)
		}(i8, u8)

		i8++
		u8++
	}

	wg.Wait()
}

func TestFixedArray(t *testing.T) {
	c := check.New(t)

	in := [3]int32{1, 2, 3}
	out := [3]int32{}

	trampoline(
		c, in, &out,
		[]byte("/\x00\x00\x00\x01/\x00\x00\x00\x02/\x00\x00\x00\x03/"))
}

func TestVariableArray(t *testing.T) {
	c := check.New(t)

	in := [3]string{"test", "something", "fun"}
	out := [3]string{}

	trampoline(
		c, in, &out,
		[]byte("/test/something/fun/"))
}

func TestArrayPtrs(t *testing.T) {
	c := check.New(t)

	in := [2]*uint32{new(uint32), new(uint32)}
	out := [2]*uint32{}

	_, err := Marshal(in, nil)
	c.Error(err)

	_, err = Unmarshal([]byte("/\x00\x00\x00\x01/\x00\x00\x00\x02/"), &out)
	c.Error(err)
}

func TestSeparator(t *testing.T) {
	c := check.New(t)

	pie := Pie{
		A: "apple",
		B: 1,
		C: "pumpkin",
		D: 2,
		E: "berry",
		F: 3,
	}

	s := Separator('#')

	p, err := s.Marshal(pie, nil)
	c.MustNotError(err)

	p2 := Pie{}
	_, err = s.Unmarshal(p, &p2)
	c.MustNotError(err)
	c.Equal(pie, p2)
}

func TestUnsupportedMarshalTypes(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		make(chan int),
		[]int{1, 2, 3},
		struct{ Ch chan int }{},
		[...]chan struct{}{make(chan struct{}), make(chan struct{})},
	}

	for i, t := range tests {
		_, err := Marshal(t, nil)
		c.Error(err, "%d did not error", i)
	}
}

func TestUnsupportedUnmarshalTypes(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		&[]int{1, 2, 3},
		&[...]int{1, 2, 3},
		&struct{ Ch chan int }{},
		&[...]chan struct{}{make(chan struct{}), make(chan struct{})},
	}

	for i, t := range tests {
		_, err := Unmarshal([]byte("/test/"), t)
		c.Error(err, "%d did not error", i)
	}
}

func TestInvalidString(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		"string/with/slashes",
		[]byte("string/with/slashes"),
	}

	for i, t := range tests {
		_, err := Marshal(t, nil)
		c.Error(err, "%d did not error", i)
	}
}

func TestTruncated(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		new(int8),
		new(int16),
		new(int32),
		new(int64),
		new(uint8),
		new(uint16),
		new(uint32),
		new(uint64),
		new(float32),
		new(float64),
		new(complex64),
		new(complex128),
		struct{ A uint64 }{},
	}

	for i, t := range tests {
		_, err := Unmarshal([]byte("/"), t)
		c.Error(err, "%d did not fail", i)
	}
}

func TestEverythingTruncated(t *testing.T) {
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

	b := MustMarshal(e, nil)
	for i := range b {
		e2 := Everything{}

		_, err := Unmarshal(b[:i], &e2)
		c.Error(err, "%d did not fail", i)
	}
}

func TestUnmarshalWithoutPointer(t *testing.T) {
	c := check.New(t)
	v := struct{ A int }{}

	_, err := Unmarshal(nil, v)
	c.Error(err)
}

func TestWrongTag(t *testing.T) {
	c := check.New(t)

	in := struct {
		_ Static `path:"in"`
	}{}
	out := struct {
		_ Static `path:"out"`
	}{}

	b, err := Marshal(in, nil)
	c.MustNotError(err)

	_, err = Unmarshal(b, &out)
	c.Error(err)
}

func TestErrors(t *testing.T) {
	c := check.New(t)

	pie := Pie{
		A: "apple",
		B: 1,
		C: "pumpkin",
		D: 2,
		E: "berry",
		F: 3,
	}

	p, err := Marshal(pie, nil)
	c.MustNotError(err)

	err = nil
	p2 := Pie{}
	for i := range p {
		_, err = Unmarshal(p[0:i], &p2)
		c.Error(err)
	}

	_, err = Unmarshal(p, &p2)
	c.NotError(err)
}

func TestMarshalIndirectTypes(t *testing.T) {
	c := check.New(t)

	path := []byte("/\x00\x00\x00\x01/")
	arrayPath := []byte("/\x00\x00\x00\x01/\x00\x00\x00\x02/")

	type ptr *int32
	var p ptr = new(int32)
	_, err := Marshal(p, nil)
	c.Error(err)

	pa := [...]ptr{new(int32), new(int32)}
	*pa[0] = 1
	*pa[1] = 2
	_, err = Marshal(pa, nil)
	c.Error(err)

	type direct int32
	var d direct = 1
	b := MustMarshal(d, nil)
	c.Equal(path, b)

	pd := [...]direct{1, 2}
	b = MustMarshal(pd, nil)
	c.Equal(arrayPath, b)
}

func TestUnmarshalIndirectTypes(t *testing.T) {
	c := check.New(t)

	path := []byte("/\x00\x00\x00\x01/")
	arrayPath := []byte("/\x00\x00\x00\x01/\x00\x00\x00\x02/")

	type ptr *int32
	var p ptr = new(int32)
	_, err := Unmarshal(path, &p)
	c.Error(err)

	pa := [...]ptr{new(int32), new(int32)}
	_, err = Unmarshal(arrayPath, &pa)
	c.Error(err)

	type direct int32
	var d direct
	MustUnmarshal(path, &d)
	c.Equal(1, d)

	pd := [...]direct{1, 2}
	MustUnmarshal(arrayPath, &pd)
	c.Equal(pd[0], 1)
	c.Equal(pd[1], 2)
}

func TestIndirectTypesCoverage(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		b(false),
		i8(1),
		i16(1),
		i32(1),
		i64(1),
		u8(1),
		u16(1),
		u32(1),
		u64(1),
		f32(1),
		f64(1),
		c64(1),
		c128(1),
	}

	for i, test := range tests {
		pb, err := Marshal(test, nil)
		c.MustNotError(err, "direct failed at %d", i)

		_, err = Unmarshal(pb, reflect.New(reflect.TypeOf(test)).Interface())
		c.MustNotError(err, "direct failed at %d", i)
	}
}

func TestExpectTagCoverage(t *testing.T) {
	c := check.New(t)

	d := DefSep.NewDecoder([]byte("/hai/"))
	d = d.ExpectTag("test")
	c.Error(d.Err)

	d = DefSep.NewDecoder([]byte("/hai/"))
	d = d.ExpectTagBytes([]byte("test"))
	c.Error(d.Err)

	d = DefSep.NewDecoder([]byte("/hai"))
	d = d.ExpectTagBytes([]byte("test"))
	c.Error(d.Err)
}

func TestExpectBytesCoverage(t *testing.T) {
	c := check.New(t)

	d := DefSep.NewDecoder([]byte("/hai"))
	d = d.ExpectBytes(nil)
	c.Error(d.Err)

	ba := [3]uint8{}
	d = DefSep.NewDecoder([]byte("/\x01\x02\x03/"))
	d = d.ExpectByteArray(ba[:])
	c.MustNotError(d.Err)
	c.Equal([3]uint8{1, 2, 3}, ba)

	ba = [3]uint8{}
	d = DefSep.NewDecoder([]byte("/\x01\x02"))
	d = d.ExpectByteArray(ba[:])
	c.Error(d.Err)
}

func BenchmarkMarshal(b *testing.B) {
	b.ReportAllocs()

	pie := Pie{
		A: "apple",
		B: 1,
		C: "apple",
		D: 1,
		E: "apple",
		F: 1,
	}

	var out []byte
	buff := make([]byte, 0, 1024)
	for i := 0; i < b.N; i++ {
		out = MustMarshal(pie, buff)
	}

	b.SetBytes(int64(len(out)))
}

func BenchmarkMarshalPathInterface(b *testing.B) {
	b.ReportAllocs()

	pie := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}

	var out []byte
	buff := make([]byte, 0, 1024)
	for i := 0; i < b.N; i++ {
		out = MustMarshal(pie, buff)
	}

	b.SetBytes(int64(len(out)))
}

func BenchmarkMarshalPathRaw(b *testing.B) {
	b.ReportAllocs()

	c := check.New(b)

	pie := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}

	var out []byte
	buff := make([]byte, 0, 1024)
	for i := 0; i < b.N; i++ {
		e := pie.MarshalPath(DefSep.NewEncoder(buff))
		c.MustNotError(e.Err)
		out = e.B
	}

	b.SetBytes(int64(len(out)))
}

func BenchmarkUnmarshal(b *testing.B) {
	b.ReportAllocs()

	pie := Pie{
		A: "apple",
		B: 1,
		C: "apple",
		D: 1,
		E: "apple",
		F: 1,
	}

	pb := MustMarshal(pie, nil)
	b.SetBytes(int64(len(pb)))

	for i := 0; i < b.N; i++ {
		p2 := Pie{}
		MustUnmarshal(pb, &p2)
	}
}

func BenchmarkUnmarshalPath(b *testing.B) {
	b.ReportAllocs()

	pie := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}

	pb := MustMarshal(pie, nil)
	b.SetBytes(int64(len(pb)))

	for i := 0; i < b.N; i++ {
		p2 := PieWithInterface{}
		MustUnmarshal(pb, &p2)
	}
}

func BenchmarkUnmarshalPathRaw(b *testing.B) {
	b.ReportAllocs()

	c := check.New(b)

	pie := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}

	pb := MustMarshal(pie, nil)
	b.SetBytes(int64(len(pb)))

	d := DefSep.NewDecoder(pb)
	c.MustNotError(d.Err)

	for i := 0; i < b.N; i++ {
		p2 := PieWithInterface{}
		dd := p2.UnmarshalPath(d)
		c.MustNotError(dd.Err)
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

	// This path isn't human-readable, so printing it anywhere is pointless
	path := MustMarshal(v, nil)

	// Reset everything, just in case
	v.Type = ""
	v.Index = 0
	v.Desirability = 0

	MustUnmarshal(path, &v)
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
		_     Static `path:"pie"` // Prefix this path with "/pie/"
		Type  string
		Taste string
	}{
		Type:  "apple",
		Taste: "fantastic",
	}

	path := MustMarshal(v, nil)
	fmt.Println(string(path))

	// Output:
	// /pie/apple/fantastic/
}
