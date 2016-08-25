package path

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/iheartradio/cog"
)

// Marshaler is the interface implemented by objects that can marshal themselves
// into a valid path
type Marshaler interface {
	MarshalPath(e Encoder) Encoder
}

// An Encoder is used for Marshaling paths. Never create this directly; use
// NewEncoder instead.
type Encoder struct {
	State
}

// NewEncoder creates a new Encoder, adding the separator as the first byte.
func (s Separator) NewEncoder(b []byte) (enc Encoder) {
	enc.State = State{
		B: b[0:0],
		s: byte(s),
	}

	enc = enc.EmitSep()

	return
}

// Marshal turns the given struct into a structured path
func (s Separator) Marshal(v interface{}, cache []byte) ([]byte, error) {
	e := s.NewEncoder(cache).Marshal(v)
	return e.B, e.Err
}

// MustMarshal is like Marshal, except it panics on failure
func (s Separator) MustMarshal(v interface{}, cache []byte) []byte {
	b, err := s.Marshal(v, cache)
	cog.Must(err, "marshal failed")
	return b
}

// Must ensures that there were no Marshaling errors
func (e Encoder) Must() []byte {
	cog.Must(e.Err, "marshal failed")
	return e.B
}

// Marshal marshals a new value in the current state
func (e Encoder) Marshal(v interface{}) Encoder {
	switch v := v.(type) {
	case Marshaler:
		return v.MarshalPath(e)

	case bool:
		return e.EmitBool(v)

	case int8:
		return e.EmitInt8(v)
	case int16:
		return e.EmitInt16(v)
	case int32:
		return e.EmitInt32(v)
	case int64:
		return e.EmitInt64(v)

	case uint8:
		return e.EmitUint8(v)
	case uint16:
		return e.EmitUint16(v)
	case uint32:
		return e.EmitUint32(v)
	case uint64:
		return e.EmitUint64(v)

	case float32:
		return e.EmitFloat32(v)
	case float64:
		return e.EmitFloat64(v)
	case complex64:
		return e.EmitComplex64(v)
	case complex128:
		return e.EmitComplex128(v)

	case string:
		return e.EmitString(v)
	case []byte:
		return e.EmitBytes(v)

	default:
		return e.marshalReflect(v)
	}
}

// EmitSep writes the path separator to the output if !s.DisableSep
func (e Encoder) EmitSep() Encoder {
	e.B = append(e.B, e.s)
	return e
}

// EmitBool writes an encoded boolean value into the buffer
func (e Encoder) EmitBool(v bool) Encoder {
	b := byte('\x00')
	if v {
		b = '\x01'
	}

	e.B = append(e.B, b)
	return e.EmitSep()
}

// EmitUint8 writes a uint8 to the output
func (e Encoder) EmitUint8(v uint8) Encoder {
	e.B = append(e.B, v)
	return e.EmitSep()
}

// EmitUint16 writes a uint16 to the output
func (e Encoder) EmitUint16(v uint16) Encoder {
	e.B = append(e.B,
		byte(v>>8),
		byte(v))
	return e.EmitSep()
}

// EmitUint32 writes a uint32 to the output
func (e Encoder) EmitUint32(v uint32) Encoder {
	e.B = append(e.B,
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v))
	return e.EmitSep()
}

// EmitUint64 writes a uint64 to the output
func (e Encoder) EmitUint64(v uint64) Encoder {
	e.B = append(e.B,
		byte(v>>56),
		byte(v>>48),
		byte(v>>40),
		byte(v>>32),
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v))
	return e.EmitSep()
}

// EmitInt8 writes an int8 to the output
func (e Encoder) EmitInt8(v int8) Encoder {
	return e.EmitUint8(uint8(v))
}

// EmitInt16 writes an int16 to the output
func (e Encoder) EmitInt16(v int16) Encoder {
	return e.EmitUint16(uint16(v))
}

// EmitInt32 writes an int32 to the output
func (e Encoder) EmitInt32(v int32) Encoder {
	return e.EmitUint32(uint32(v))
}

// EmitInt64 writes an int64 to the output
func (e Encoder) EmitInt64(v int64) Encoder {
	return e.EmitUint64(uint64(v))
}

// EmitFloat32 writes a float32 to the output
func (e Encoder) EmitFloat32(v float32) Encoder {
	return e.EmitUint32(math.Float32bits(v))
}

// EmitFloat64 writes a float64 to the output
func (e Encoder) EmitFloat64(v float64) Encoder {
	return e.EmitUint64(math.Float64bits(v))
}

// EmitComplex64 writes a complex64 to the output
func (e Encoder) EmitComplex64(v complex64) Encoder {
	rl := math.Float32bits(real(v))
	ig := math.Float32bits(imag(v))

	return e.EmitUint64(uint64(rl)<<32 | uint64(ig))
}

// EmitComplex128 writes a complex128 to the output
func (e Encoder) EmitComplex128(v complex128) Encoder {
	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, math.Float64bits(real(v)))
	e.B = append(e.B, b...)

	binary.BigEndian.PutUint64(b, math.Float64bits(imag(v)))
	e.B = append(e.B, b...)

	return e.EmitSep()
}

// EmitString writes an encoded string value into the buffer
func (e Encoder) EmitString(v string) Encoder {
	if strings.IndexByte(v, e.s) != -1 {
		e.Err = fmt.Errorf("path: invalid string: may not contain a `%c`", e.s)
		return e
	}

	e.B = append(e.B, v...)
	return e.EmitSep()
}

// EmitBytes writes an encoded byte slice value into the buffer
func (e Encoder) EmitBytes(v []byte) Encoder {
	if bytes.IndexByte(v, e.s) != -1 {
		e.Err = fmt.Errorf("path: invalid []byte: may not contain a `%c`", e.s)
		return e
	}

	e.B = append(e.B, v...)
	return e.EmitSep()
}

func (e Encoder) marshalReflect(v interface{}) Encoder {
	rv := reflect.ValueOf(v)

	var rt reflect.Type
	if rv.IsValid() {
		rt = rv.Type()
	}

	// A nil obviously can't be a path
	if rt == nil {
		e.Err = fmt.Errorf("path: a nil cannot be turned into a path")
		return e
	}

	kind := rt.Kind()
	if kind == reflect.Ptr {
		e.Err = fmt.Errorf("path: sorry, pointers are not allowed")
		return e
	}

	switch kind {
	case reflect.Struct:
		return e.marshalStruct(rt, rv)

	case reflect.Array:
		return e.marshalArray(rt, rv)

	case reflect.Bool:
		return e.EmitBool(rv.Bool())

	case reflect.Int8:
		return e.EmitInt8(int8(rv.Int()))
	case reflect.Int16:
		return e.EmitInt16(int16(rv.Int()))
	case reflect.Int32:
		return e.EmitInt32(int32(rv.Int()))
	case reflect.Int64:
		return e.EmitInt64(rv.Int())

	case reflect.Uint8:
		return e.EmitUint8(uint8(rv.Uint()))
	case reflect.Uint16:
		return e.EmitUint16(uint16(rv.Uint()))
	case reflect.Uint32:
		return e.EmitUint32(uint32(rv.Uint()))
	case reflect.Uint64:
		return e.EmitUint64(rv.Uint())

	case reflect.Float32:
		return e.EmitFloat32(float32(rv.Float()))
	case reflect.Float64:
		return e.EmitFloat64(rv.Float())
	case reflect.Complex64:
		return e.EmitComplex64(complex64(rv.Complex()))
	case reflect.Complex128:
		return e.EmitComplex128(rv.Complex())

	default:
		e.Err = fmt.Errorf("path: unsupported marshal kind: %s", kind)
		return e
	}
}

func (e Encoder) marshalStruct(rt reflect.Type, rv reflect.Value) Encoder {
	n := rv.NumField()
	for i := 0; i < n; i++ {
		f := rv.Field(i)

		if f.Type() == staticType {
			tag := rt.Field(i).Tag.Get("path")
			e.B = append(e.B, tag...)
			e.B = append(e.B, e.s)
		} else {
			if !f.CanInterface() { // unexported
				continue
			}

			e = e.Marshal(f.Interface())
		}

		if e.Err != nil {
			return e
		}
	}

	return e
}

func (e Encoder) marshalArray(rt reflect.Type, rv reflect.Value) Encoder {
	n := rv.Len()
	for i := 0; i < n; i++ {
		e = e.Marshal(rv.Index(i).Interface())
		if e.Err != nil {
			break
		}
	}

	return e
}
