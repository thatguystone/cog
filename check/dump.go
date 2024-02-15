package check

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

type circularKey struct {
	t   reflect.Type
	ptr unsafe.Pointer
}

func makeCircularKey(rv reflect.Value) (key circularKey, ok bool) {
	switch rv.Kind() {
	default:
		return

	case reflect.Pointer, reflect.Map, reflect.Slice:
		if rv.IsNil() {
			return
		}

		key = circularKey{rv.Type(), rv.UnsafePointer()}
		ok = true
		return
	}
}

type dumper struct {
	buf         bytes.Buffer
	indentDepth int
	seen        map[circularKey]struct{}
	ids         map[circularKey]int
}

func dump(v any, initialIndent int) string {
	d := dumper{
		indentDepth: initialIndent,
		seen:        make(map[circularKey]struct{}),
		ids:         make(map[circularKey]int),
	}

	if initialIndent > 0 {
		d.writeIndent()
	}

	if v == nil {
		d.buf.WriteString("nil")
	} else {
		rv := reflect.ValueOf(v)
		d.walkCirculars(rv)
		d.fmtVal(rv)
	}

	return d.buf.String()
}

const (
	dumpIndent   = "    "
	maxBase10Len = 26 // len("-9_223_372_036_854_775_808")
	maxBase16Len = 19 // len("-0x8000000000000000")
)

func (d *dumper) walkCirculars(rv reflect.Value) {
	mightCircular := func(rt reflect.Type) bool {
		switch rt.Kind() {
		case reflect.Pointer:
		case reflect.Interface:
		case reflect.Array:
		case reflect.Slice:
		case reflect.Map:
		case reflect.Struct:
		default:
			return false
		}

		return true
	}

	key, ok := makeCircularKey(rv)
	if ok {
		if _, ok := d.seen[key]; ok {
			if _, ok := d.ids[key]; !ok {
				// Start ids at 1, not 0, to avoid nil-ptr-looking values, eg.
				// ptr0
				d.ids[key] = len(d.ids) + 1
			}

			return
		}

		d.seen[key] = struct{}{}
		defer delete(d.seen, key)
	}

	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		d.walkCirculars(rv.Elem())
	case reflect.Array, reflect.Slice:
		if mightCircular(rv.Type().Elem()) {
			for i := range rv.Len() {
				d.walkCirculars(rv.Index(i))
			}
		}
	case reflect.Map:
		rt := rv.Type()
		if mightCircular(rt.Key()) || mightCircular(rt.Elem()) {
			for iter := rv.MapRange(); iter.Next(); {
				d.walkCirculars(iter.Key())
				d.walkCirculars(iter.Value())
			}
		}
	case reflect.Struct:
		for i := range rv.NumField() {
			d.walkCirculars(rv.Field(i))
		}
	}
}

func (d *dumper) fmtVal(rv reflect.Value) {
	key, ok := makeCircularKey(rv)
	if ok {
		if _, ok := d.seen[key]; ok {
			if rv.Kind() == reflect.Pointer {
				d.buf.WriteByte('(')
			}

			d.writeType(rv)

			if rv.Kind() == reflect.Pointer {
				d.buf.WriteByte(')')
			}

			fmt.Fprintf(&d.buf, "(0x%x)", d.ids[key])
			return
		}

		if id, ok := d.ids[key]; ok {
			fmt.Fprintf(&d.buf, "/* 0x%x */", id)
		}

		d.seen[key] = struct{}{}
		defer delete(d.seen, key)
	}

	d.writeAnnotation(rv)

	switch rv.Kind() {
	case reflect.Bool:
		d.fmtBool(rv)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.fmtInt(rv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d.fmtUint(rv)
	case reflect.Float32, reflect.Float64:
		d.fmtFloat(rv)
	case reflect.Complex64, reflect.Complex128:
		d.fmtComplex(rv)
	case reflect.String:
		d.fmtString(rv)
	case reflect.Array, reflect.Slice:
		d.fmtSlice(rv)
	case reflect.Map:
		d.fmtMap(rv)
	case reflect.Struct:
		d.fmtStruct(rv)
	case reflect.Pointer:
		d.fmtPointer(rv)
	case reflect.Interface:
		d.fmtInterface(rv)
	case reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer:
		d.fmtOpaquePointer(rv)
	default:
		panic(fmt.Errorf("unexpected type: %s", rv.Type()))
	}
}

func (d *dumper) fmtBool(rv reflect.Value) {
	var (
		typeName = rv.Type().String()
		hasType  = typeName != rv.Kind().String()
	)

	if hasType {
		d.buf.WriteString(typeName)
		d.buf.WriteByte('(')
	}

	if rv.Bool() {
		d.buf.WriteString("true")
	} else {
		d.buf.WriteString("false")
	}

	if hasType {
		d.buf.WriteByte(')')
	}
}

func (d *dumper) fmtInt(rv reflect.Value) {
	d.writeType(rv)
	d.buf.WriteByte('(')

	d.buf.Grow(maxBase10Len)
	b := d.buf.AvailableBuffer()
	b = strconv.AppendInt(b, rv.Int(), 10)
	b = fmtBase10(b)
	d.buf.Write(b)

	d.buf.WriteByte(')')
}

func (d *dumper) fmtUint(rv reflect.Value) {
	d.writeType(rv)
	d.buf.WriteByte('(')

	d.buf.Grow(maxBase10Len)
	b := d.buf.AvailableBuffer()
	b = strconv.AppendUint(b, rv.Uint(), 10)
	b = fmtBase10(b)
	d.buf.Write(b)

	d.buf.WriteByte(')')
}

func (d *dumper) fmtFloat(rv reflect.Value) {
	d.writeType(rv)
	d.buf.WriteByte('(')
	d.writeFloat(rv.Float(), true)
	d.buf.WriteByte(')')
}

func (d *dumper) fmtComplex(rv reflect.Value) {
	d.writeType(rv)
	d.buf.WriteByte('(')

	v := rv.Complex()
	d.writeFloat(real(v), false)

	var (
		im   = imag(v)
		sign = " + "
	)

	if im < 0 {
		im = -im
		sign = " - "
	}

	d.buf.WriteString(sign)
	d.writeFloat(im, false)
	d.buf.WriteByte('i')

	d.buf.WriteByte(')')
}

func (d *dumper) fmtString(rv reflect.Value) {
	var (
		typeName = rv.Type().String()
		hasType  = typeName != rv.Kind().String()
	)

	if hasType {
		d.buf.WriteString(typeName)
		d.buf.WriteByte('(')
	}

	d.writeGoString(rv.String())

	if hasType {
		d.buf.WriteByte(')')
	}
}

func (d *dumper) fmtSlice(rv reflect.Value) {
	d.writeType(rv)

	if rv.Kind() == reflect.Slice {
		if rv.IsNil() {
			d.buf.WriteString("(nil)")
			return
		}

		if rv.Len() == 0 {
			d.buf.WriteString("{}")
			return
		}
	}

	d.buf.WriteString("{")
	d.indent()

	if rv.Type().Elem() == reflect.TypeOf(byte(0)) {
		var (
			n            = rv.Len()
			nlines       = (n / 8) + 1
			lineOverhead = (len(dumpIndent) * d.indentDepth) + 1 // indent + nl
		)

		d.buf.Grow((n * len("0x00, ")) + (nlines * lineOverhead))

		for i := range n {
			if i%8 == 0 {
				d.buf.WriteString("\n")
				d.writeIndent()
			} else {
				d.buf.WriteByte(' ')
			}

			v := rv.Index(i).Uint()
			b := d.buf.AvailableBuffer()
			b = append(b, "0x"...)
			if v < 0x10 {
				b = append(b, '0')
			}
			b = strconv.AppendUint(b, v, 16)
			b = append(b, ',')
			d.buf.Write(b)
		}

		d.buf.WriteString("\n")
	} else {
		d.buf.WriteString("\n")
		for i := range rv.Len() {
			d.writeIndent()
			d.fmtVal(rv.Index(i))
			d.buf.WriteString(",\n")
		}
	}

	d.dedent()
	d.writeIndent()
	d.buf.WriteByte('}')
}

func (d *dumper) fmtMap(rv reflect.Value) {
	d.writeType(rv)

	if rv.IsNil() {
		d.buf.WriteString("(nil)")
		return
	}

	if rv.Len() == 0 {
		d.buf.WriteString("{}")
		return
	}

	d.buf.WriteString("{\n")
	d.indent()

	for _, kv := range sortMap(rv) {
		d.writeIndent()
		d.fmtVal(kv.k)
		d.buf.WriteString(": ")
		d.fmtVal(kv.v)
		d.buf.WriteString(",\n")
	}

	d.dedent()
	d.writeIndent()
	d.buf.WriteByte('}')
}

func (d *dumper) fmtStruct(rv reflect.Value) {
	d.writeType(rv)

	var (
		rt       = rv.Type()
		numField = rt.NumField()
	)

	if numField == 0 {
		d.buf.WriteString("{}")
		return
	}

	d.buf.WriteString("{\n")
	d.indent()

	for i := range numField {
		d.writeIndent()
		d.buf.WriteString(rt.Field(i).Name)
		d.buf.WriteString(": ")
		d.fmtVal(rv.Field(i))
		d.buf.WriteString(",\n")
	}

	d.dedent()
	d.writeIndent()
	d.buf.WriteByte('}')
}

func (d *dumper) fmtPointer(rv reflect.Value) {
	if rv.IsNil() {
		d.buf.WriteByte('(')
		d.writeType(rv)
		d.buf.WriteByte(')')
		d.buf.WriteString("(nil)")
		return
	}

	typeChange := rv.Type().Name() != ""
	if typeChange {
		d.writeType(rv)
		d.buf.WriteByte('(')
	}

	d.buf.WriteByte('&')
	d.fmtVal(rv.Elem())

	if typeChange {
		d.buf.WriteByte(')')
	}
}

func (d *dumper) fmtInterface(rv reflect.Value) {
	d.writeType(rv)
	d.buf.WriteByte('(')

	if rv.IsNil() {
		d.buf.WriteString("nil")
	} else {
		d.fmtVal(rv.Elem())
	}

	d.buf.WriteByte(')')
}

func (d *dumper) fmtOpaquePointer(rv reflect.Value) {
	d.buf.WriteByte('(')
	d.writeType(rv)
	d.buf.WriteString(")(")

	var ptr uint64
	if rv.Kind() == reflect.Uintptr {
		ptr = rv.Uint()
	} else {
		ptr = uint64(rv.Pointer())
	}

	if ptr == 0 {
		d.buf.WriteString("nil")
	} else {
		d.buf.Grow(maxBase16Len)
		b := d.buf.AvailableBuffer()
		b = append(b, "0x"...)
		b = strconv.AppendUint(b, ptr, 16)
		d.buf.Write(b)
	}

	d.buf.WriteByte(')')
}

func (d *dumper) writeAnnotation(rv reflect.Value) {
	// Only annotate concrete values: pointers and interfaces all resolve into
	// concrete types, so annotating them results in printing the same thing
	// multiple times
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		return
	}

	if !rv.CanInterface() {
		tmp, ok := forceCanInterface(rv)
		if !ok {
			return
		}

		rv = tmp
	}

	// rv can't be a ptr at this point, but methods might have ptr receivers
	if rv.CanAddr() {
		rv = rv.Addr()
	}

	str, ok := func() (str string, ok bool) {
		defer func() {
			if r := recover(); r != nil {
				d.buf.WriteString("/* ")
				fmt.Fprintf(&d.buf, "(PANIC=%q)", r)
				d.buf.WriteString(" */")
			}
		}()

		switch v := rv.Interface().(type) {
		case error:
			str = v.Error()
		case fmt.Stringer:
			str = v.String()
		default:
			return
		}

		ok = true
		return
	}()

	if !ok {
		return
	}

	d.buf.WriteString("/* ")
	d.writeGoString(str)
	d.buf.WriteString(" */")
}

func (d *dumper) writeGoString(v string) {
	d.buf.Grow(1 + len(v) + 1)

	b := d.buf.AvailableBuffer()

	backquote := strings.Contains(v, `"`) &&
		!strings.ContainsFunc(v, func(r rune) bool { return !unicode.IsPrint(r) }) &&
		strconv.CanBackquote(v)
	if backquote {
		b = append(b, '`')
		b = append(b, v...)
		b = append(b, '`')
	} else {
		b = strconv.AppendQuote(b, v)
	}

	d.buf.Write(b)
}

func (d *dumper) writeType(rv reflect.Value) {
	name := rv.Type().String()
	name = strings.ReplaceAll(name, "interface {}", "any")
	name = strings.ReplaceAll(name, "interface{}", "any")
	d.buf.WriteString(name)
}

func (d *dumper) writeFloat(v float64, ensureDot bool) {
	var (
		prec = -1
		verb = byte('g')
	)

	if ensureDot && math.Trunc(v) == v {
		// Force a ".0" to disambiguate int==1 and float==1.0
		verb = 'f'
		prec = 1
	}

	d.buf.Grow(32)
	b := d.buf.AvailableBuffer()
	b = strconv.AppendFloat(b, v, verb, prec, 64)
	d.buf.Write(b)
}

func (d *dumper) indent() {
	d.indentDepth++
}

func (d *dumper) dedent() {
	d.indentDepth--
}

func (d *dumper) writeIndent() {
	d.buf.Grow(len(dumpIndent) * d.indentDepth)
	for range d.indentDepth {
		d.buf.WriteString(dumpIndent)
	}
}

func fmtBase10(s []byte) []byte {
	ps := s
	if ps[0] == '-' {
		ps = ps[1:]
	}

	var (
		n    = len(ps)
		nsep = (n - 1) / 3
		ret  = s[:len(s)+nsep]
		reti = len(ret) - 1
	)

	prepend := func(b byte) {
		ret[reti] = b
		reti--
	}

	for i := range ps {
		if i > 0 && i%3 == 0 {
			prepend('_')
		}
		prepend(s[len(s)-1-i])
	}

	return ret
}
