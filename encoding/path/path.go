// Package path provides Marshaling and Unmarshaling for [ ]byte-encoded paths.
//
// For example, if you have a path like "/pies/apple/0/", of the standard form
// "/pies/<Type:string>/<Index:uint32>/", this helps you Unmarshal that into a
// Pies struct with the fields "Type" and "Index", and vis-versa.
//
// This encodes into a binary, non-human-readable format when using anything but
// strings. It also only supports Marshaling and Unmarshaling into structs using
// primitive types (with fixed byte sizes), which means that your paths must be
// strictly defined before using this.
//
// Warning: do not use this for anything that will change over time and that
// needs to remain backwards compatible. For that, you should use something like
// capnproto.
package path

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/thatguystone/cog"
)

// Separator allows you to change the path separator used.
type Separator byte

// Static is used to place a static, unchanging element into the path.
type Static struct{}

// State is the state of the Marshaler or Unmarshaler.
type State struct {
	// Where bytes are appended to and removed from
	B []byte

	// Any error encountered along the way
	Err error

	// If a separator is definitely needed
	NeedSep bool

	// If no separators should be emitted
	DisableSep bool

	s byte
}

// Marshaler is the interface implemented by objects that can marshal themselves
// into a valid path
type Marshaler interface {
	MarshalPath(s State) State
}

// Unmarshaler is the interface implemented by objects that can unmarshal
// themselves from a bytes.Buffer.
type Unmarshaler interface {
	UnmarshalPath(s State) State
}

var staticType = reflect.TypeOf(Static{})

// Marshal turns the given struct into a structured path
func (sep Separator) Marshal(v interface{}, cache []byte) ([]byte, error) {
	s := State{
		B:       cache,
		NeedSep: true,
		s:       byte(sep),
	}

	s = s.Marshal(v)

	return s.B, s.Err
}

// MustMarshal is like Marshal, except it panics on failure
func (sep Separator) MustMarshal(v interface{}, cache []byte) []byte {
	b, err := sep.Marshal(v, cache)
	cog.Must(err, "marshal failed")
	return b
}

// Marshal marshals a new value in the current state
func (s State) Marshal(v interface{}) State {
	s = s.MaybeEmitSep()

	switch v := v.(type) {
	case Marshaler:
		return v.MarshalPath(s)

	case bool:
		return s.EmitBool(v)
	case *bool:
		return s.EmitBool(*v)

	case int8:
		return s.EmitInt8(v)
	case *int8:
		return s.EmitInt8(*v)

	case int16:
		return s.EmitInt16(v)
	case *int16:
		return s.EmitInt16(*v)

	case int32:
		return s.EmitInt32(v)
	case *int32:
		return s.EmitInt32(*v)

	case int64:
		return s.EmitInt64(v)
	case *int64:
		return s.EmitInt64(*v)

	case uint8:
		return s.EmitUint8(v)
	case *uint8:
		return s.EmitUint8(*v)

	case uint16:
		return s.EmitUint16(v)
	case *uint16:
		return s.EmitUint16(*v)

	case uint32:
		return s.EmitUint32(v)
	case *uint32:
		return s.EmitUint32(*v)

	case uint64:
		return s.EmitUint64(v)
	case *uint64:
		return s.EmitUint64(*v)

	case float32:
		return s.EmitFloat32(v)
	case *float32:
		return s.EmitFloat32(*v)

	case float64:
		return s.EmitFloat64(v)
	case *float64:
		return s.EmitFloat64(*v)

	case complex64:
		return s.EmitComplex64(v)
	case *complex64:
		return s.EmitComplex64(*v)

	case complex128:
		return s.EmitComplex128(v)
	case *complex128:
		return s.EmitComplex128(*v)

	case string:
		return s.EmitString(v)
	case *string:
		return s.EmitString(*v)

	case []byte:
		return s.EmitBytes(v)
	case *[]byte:
		return s.EmitBytes(*v)

	default:
		return s.marshalReflect(reflect.ValueOf(v))
	}
}

// EmitSep writes the path separator to the output if !s.DisableSep
func (s State) EmitSep() State {
	if !s.DisableSep {
		s.B = append(s.B, s.s)
		s.NeedSep = false
	}

	return s
}

// MaybeEmitSep writes the path separator to the output if s.NeedSep && !s.DisableSep
func (s State) MaybeEmitSep() State {
	if s.NeedSep && !s.DisableSep {
		s = s.EmitSep()
	}

	return s
}

// EmitBool writes an encoded boolean value into the buffer
func (s State) EmitBool(v bool) State {
	b := byte('\x00')
	if v {
		b = '\x01'
	}

	s = s.MaybeEmitSep()
	s.B = append(s.B, b)
	return s.EmitSep()
}

// EmitUint8 writes a uint8 to the output
func (s State) EmitUint8(v uint8) State {
	s = s.MaybeEmitSep()
	s.B = append(s.B, v)
	return s.EmitSep()
}

// EmitUint16 writes a uint16 to the output
func (s State) EmitUint16(v uint16) State {
	s = s.MaybeEmitSep()
	s.B = append(s.B,
		byte(v>>8),
		byte(v))
	return s.EmitSep()
}

// EmitUint32 writes a uint32 to the output
func (s State) EmitUint32(v uint32) State {
	s = s.MaybeEmitSep()
	s.B = append(s.B,
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v))
	return s.EmitSep()
}

// EmitUint64 writes a uint64 to the output
func (s State) EmitUint64(v uint64) State {
	s = s.MaybeEmitSep()
	s.B = append(s.B,
		byte(v>>56),
		byte(v>>48),
		byte(v>>40),
		byte(v>>32),
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v))
	return s.EmitSep()
}

// EmitInt8 writes an int8 to the output
func (s State) EmitInt8(v int8) State {
	return s.EmitUint8(uint8(v))
}

// EmitInt16 writes an int16 to the output
func (s State) EmitInt16(v int16) State {
	return s.EmitUint16(uint16(v))
}

// EmitInt32 writes an int32 to the output
func (s State) EmitInt32(v int32) State {
	return s.EmitUint32(uint32(v))
}

// EmitInt64 writes an int64 to the output
func (s State) EmitInt64(v int64) State {
	return s.EmitUint64(uint64(v))
}

// EmitFloat32 writes a float32 to the output
func (s State) EmitFloat32(v float32) State {
	return s.EmitUint32(math.Float32bits(v))
}

// EmitFloat64 writes a float64 to the output
func (s State) EmitFloat64(v float64) State {
	return s.EmitUint64(math.Float64bits(v))
}

// EmitComplex64 writes a complex64 to the output
func (s State) EmitComplex64(v complex64) State {
	rl := math.Float32bits(real(v))
	ig := math.Float32bits(imag(v))

	return s.EmitUint64(uint64(rl)<<32 | uint64(ig))
}

// EmitComplex128 writes a complex128 to the output
func (s State) EmitComplex128(v complex128) State {
	s = s.MaybeEmitSep()

	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, math.Float64bits(real(v)))
	s.B = append(s.B, b...)

	binary.BigEndian.PutUint64(b, math.Float64bits(imag(v)))
	s.B = append(s.B, b...)

	return s.EmitSep()
}

// EmitString writes an encoded string value into the buffer
func (s State) EmitString(v string) State {
	if strings.IndexByte(v, s.s) != -1 {
		s.Err = fmt.Errorf("path: invalid string: may not contain a `%c`", s.s)
		return s
	}

	s = s.MaybeEmitSep()
	s.B = append(s.B, v...)
	return s.EmitSep()
}

// EmitBytes writes an encoded byte slice value into the buffer
func (s State) EmitBytes(v []byte) State {
	if bytes.IndexByte(v, s.s) != -1 {
		s.Err = fmt.Errorf("path: invalid []byte: may not contain a `%c`", s.s)
		return s
	}

	s = s.MaybeEmitSep()
	s.B = append(s.B, v...)
	return s.EmitSep()
}

func (s State) marshalReflect(rv reflect.Value) State {
	var rt reflect.Type
	if rv.IsValid() {
		rt = rv.Type()
	}

	// A nil obviously can't be a path
	if rt == nil {
		s.Err = fmt.Errorf("path: a nil cannot be turned into a path")
		return s
	}

	kind := rt.Kind()
	if kind == reflect.Ptr {
		return s.marshalReflect(reflect.Indirect(rv))
	}

	switch kind {
	case reflect.Struct:
		return s.marshalStruct(rt, rv)

	case reflect.Array:
		return s.marshalArray(rt, rv)

	default:
		s.Err = fmt.Errorf("path: unsupported marshal kind: %s", kind)
		return s
	}
}

func (s State) marshalStruct(rt reflect.Type, rv reflect.Value) State {
	n := rv.NumField()
	for i := 0; i < n; i++ {
		f := rv.Field(i)

		if f.Type() == staticType {
			s = s.Marshal(rt.Field(i).Tag.Get("path"))
		} else {
			if !f.CanInterface() { // unexported
				continue
			}

			if f.Kind() == reflect.Ptr {
				if f.IsNil() {
					s.Err = fmt.Errorf("path: cannot Marshal from a nil pointer")
					return s
				}

				f = reflect.Indirect(f)
			}

			s = s.Marshal(f.Interface())
		}

		if s.Err != nil {
			return s
		}
	}

	return s
}

func (s State) marshalArray(rt reflect.Type, rv reflect.Value) State {
	fixed := hasFixedSize(rt.Elem().Kind())

	disabled := s.DisableSep
	s.DisableSep = fixed

	n := rv.Len()
	for i := 0; i < n; i++ {
		s = s.Marshal(rv.Index(i).Interface())
		if s.Err != nil {
			break
		}
	}

	s.DisableSep = disabled
	s.NeedSep = fixed

	return s.MaybeEmitSep()
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func (sep Separator) Unmarshal(b []byte, v interface{}) (unused []byte, err error) {
	s := State{
		B:       b,
		NeedSep: true,
		s:       byte(sep),
	}

	s = s.Unmarshal(v)
	return s.B, s.Err
}

// MustUnmarshal is like Unmarshal, except it panics on failure
func (sep Separator) MustUnmarshal(b []byte, p interface{}) (unused []byte) {
	unused, err := sep.Unmarshal(b, p)
	cog.Must(err, "unmarshal failed")
	return unused
}

// Unmarshal unmarshals from the current state into the given value
func (s State) Unmarshal(v interface{}) State {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return s
	}

	switch v := v.(type) {
	case Unmarshaler:
		return v.UnmarshalPath(s)

	case *bool:
		return s.ExpectBool(v)

	case *int8:
		return s.ExpectInt8(v)

	case *int16:
		return s.ExpectInt16(v)

	case *int32:
		return s.ExpectInt32(v)

	case *int64:
		return s.ExpectInt64(v)

	case *uint8:
		return s.ExpectUint8(v)

	case *uint16:
		return s.ExpectUint16(v)

	case *uint32:
		return s.ExpectUint32(v)

	case *uint64:
		return s.ExpectUint64(v)

	case *float32:
		return s.ExpectFloat32(v)

	case *float64:
		return s.ExpectFloat64(v)

	case *complex64:
		return s.ExpectComplex64(v)

	case *complex128:
		return s.ExpectComplex128(v)

	case *string:
		return s.ExpectString(v)

	case *[]byte:
		return s.ExpectBytes(v)

	default:
		return s.unmarshalReflect(v)
	}
}

// MaybeExpectSep looks for the delimiter as the next byte in the buffer, only
// if NeedSep is set
func (s State) MaybeExpectSep() State {
	if s.NeedSep {
		s = s.ExpectSep()
	}

	return s
}

// ExpectSep looks for the delimiter as the next byte in the buffer
func (s State) ExpectSep() State {
	if !s.DisableSep {
		if len(s.B) == 0 || s.B[0] != s.s {
			s.Err = fmt.Errorf("path: invalid path: missing delimiter")
		} else {
			s.B = s.B[1:]
			s.NeedSep = false
		}
	}

	return s
}

// ExpectBool decodes an encoded boolean value from the buffer
func (s State) ExpectBool(v *bool) State {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return s
	}

	if len(s.B) == 0 {
		s.Err = fmt.Errorf("path: invalid path: bool truncated")
		return s
	}

	*v = s.B[0] != '\x00'
	s.B = s.B[1:]
	return s.ExpectSep()
}

func (s State) readInt(size int) (int64, State) {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return 0, s
	}

	if len(s.B) < size {
		s.Err = fmt.Errorf("path: invalid path: need %d for int", size)
		return 0, s
	}

	ib := make([]byte, 8)
	copy(ib[8-size:], s.B)

	s.B = s.B[size:]
	v := (int64(ib[0]) << 56) |
		(int64(ib[1]) << 48) |
		(int64(ib[2]) << 40) |
		(int64(ib[3]) << 32) |
		(int64(ib[4]) << 24) |
		(int64(ib[5]) << 16) |
		(int64(ib[6]) << 8) |
		(int64(ib[7]))

	return v, s.ExpectSep()
}

func (s State) readUint(size int) (uint64, State) {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return 0, s
	}

	if len(s.B) < size {
		s.Err = fmt.Errorf("path: invalid path: need %d for int", size)
		return 0, s
	}

	ib := make([]byte, 8)
	copy(ib[8-size:], s.B)

	s.B = s.B[size:]
	v := (uint64(ib[0]) << 56) |
		(uint64(ib[1]) << 48) |
		(uint64(ib[2]) << 40) |
		(uint64(ib[3]) << 32) |
		(uint64(ib[4]) << 24) |
		(uint64(ib[5]) << 16) |
		(uint64(ib[6]) << 8) |
		(uint64(ib[7]))

	return v, s.ExpectSep()
}

// ExpectInt8 decodes an encoded int8 value from the buffer
func (s State) ExpectInt8(v *int8) State {
	i, s := s.readInt(1)
	if s.Err == nil {
		*v = int8(i)
	}

	return s
}

// ExpectInt16 decodes an encoded int16 value from the buffer
func (s State) ExpectInt16(v *int16) State {
	i, s := s.readInt(2)
	if s.Err == nil {
		*v = int16(i)
	}

	return s
}

// ExpectInt32 decodes an encoded int32 value from the buffer
func (s State) ExpectInt32(v *int32) State {
	i, s := s.readInt(4)
	if s.Err == nil {
		*v = int32(i)
	}

	return s
}

// ExpectInt64 decodes an encoded int64 value from the buffer
func (s State) ExpectInt64(v *int64) State {
	i, s := s.readInt(8)
	if s.Err == nil {
		*v = int64(i)
	}

	return s
}

// ExpectUint8 decodes an encoded uint8 value from the buffer
func (s State) ExpectUint8(v *uint8) State {
	i, s := s.readUint(1)
	if s.Err == nil {
		*v = uint8(i)
	}
	return s
}

// ExpectUint16 decodes an encoded uint16 value from the buffer
func (s State) ExpectUint16(v *uint16) State {
	i, s := s.readUint(2)
	if s.Err == nil {
		*v = uint16(i)
	}
	return s
}

// ExpectUint32 decodes an encoded uint32 value from the buffer
func (s State) ExpectUint32(v *uint32) State {
	i, s := s.readUint(4)
	if s.Err == nil {
		*v = uint32(i)
	}
	return s
}

// ExpectUint64 decodes an encoded uint64 value from the buffer
func (s State) ExpectUint64(v *uint64) State {
	i, s := s.readUint(8)
	if s.Err == nil {
		*v = uint64(i)
	}
	return s
}

// ExpectFloat32 decodes an encoded uint64 value from the buffer
func (s State) ExpectFloat32(v *float32) State {
	i, s := s.readUint(4)
	if s.Err == nil {
		*v = math.Float32frombits(uint32(i))
	}
	return s
}

// ExpectFloat64 decodes an encoded uint64 value from the buffer
func (s State) ExpectFloat64(v *float64) State {
	i, s := s.readUint(8)
	if s.Err == nil {
		*v = math.Float64frombits(uint64(i))
	}
	return s
}

// ExpectComplex64 decodes an encoded complex64 value from the buffer
func (s State) ExpectComplex64(v *complex64) State {
	i, s := s.readUint(8)
	if s.Err == nil {
		rl := math.Float32frombits(uint32(i >> 32))
		ig := math.Float32frombits(uint32(i))
		*v = complex(rl, ig)
	}
	return s
}

// ExpectComplex128 decodes an encoded complex128 value from the buffer
func (s State) ExpectComplex128(v *complex128) State {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return s
	}

	if s.Err == nil && len(s.B) < 16 {
		s.Err = fmt.Errorf("path: invalid path: complex128 truncated")
	}

	if s.Err != nil {
		return s
	}

	rl := math.Float64frombits(binary.BigEndian.Uint64(s.B))
	ig := math.Float64frombits(binary.BigEndian.Uint64(s.B[8:]))
	*v = complex(rl, ig)

	s.B = s.B[16:]

	return s.ExpectSep()
}

// ExpectString decodes an encoded string value from the buffer
func (s State) ExpectString(v *string) State {
	i := bytes.IndexByte(s.B, s.s)
	if i == -1 {
		s.Err = fmt.Errorf("path: failed to read string: missing separator")
		return s
	}

	*v = string(s.B[:i])
	s.B = s.B[i+1:]

	return s
}

// ExpectTag reads expects the next string it reads from the buffer to be the
// given tag
func (s State) ExpectTag(tag string) State {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return s
	}

	n := len(tag)
	if len(s.B) < n {
		s.Err = fmt.Errorf("path: failed to read tag: tag truncated")
		return s
	}

	got := string(s.B[:n])
	if got != tag {
		s.Err = fmt.Errorf("path: tag mismatch: %s != %s", got, tag)
		return s
	}

	s.B = s.B[n:]

	return s.ExpectSep()
}

// ExpectTagBytes reads expects the next byte slice it reads from the buffer to
// be the given tag. This is a faster version of ExpectTag since it requires no
// memory allocations (assuming you pass in a pre-allocated, read-only byte
// slice).
func (s State) ExpectTagBytes(tag []byte) State {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return s
	}

	n := len(tag)
	if len(s.B) < n {
		s.Err = fmt.Errorf("path: failed to read tag: tag truncated")
		return s
	}

	if !bytes.Equal(s.B[:n], tag) {
		s.Err = fmt.Errorf("path: tag mismatch: %s != %s",
			string(s.B[:n]),
			string(tag))
		return s
	}

	s.B = s.B[n:]

	return s.ExpectSep()
}

// ExpectBytes decodes an encoded byte slice value from the buffer
func (s State) ExpectBytes(v *[]byte) State {
	s = s.MaybeExpectSep()
	if s.Err != nil {
		return s
	}

	i := bytes.IndexByte(s.B, s.s)
	if i == -1 {
		s.Err = fmt.Errorf("path: failed to read []byte: missing separator")
		return s
	}

	*v = s.B[:i]
	s.B = s.B[i:]

	return s.ExpectSep()
}

func (s State) unmarshalReflect(v interface{}) State {
	rt := reflect.TypeOf(v)

	// A nil obviously can't be a path
	if rt == nil {
		s.Err = fmt.Errorf("path: cannot unmarshal into a nil value")
		return s
	}

	kind := rt.Kind()
	if kind != reflect.Ptr {
		s.Err = fmt.Errorf("path: need a pointer to unmarshal into")
		return s
	}

	rv := reflect.Indirect(reflect.ValueOf(v))
	rt = rv.Type()
	kind = rt.Kind()

	switch kind {
	case reflect.Struct:
		return s.unmarshalStruct(rt, rv)

	case reflect.Array:
		return s.unmarshalArray(rt, rv)

	default:
		s.Err = fmt.Errorf("path: unsupported unmarshal kind: %s", kind)
		return s
	}
}

func (s State) unmarshalStruct(rt reflect.Type, v reflect.Value) State {
	n := v.NumField()
	for i := 0; i < n; i++ {
		f := v.Field(i)
		ft := f.Type()

		if ft == staticType {
			s = s.ExpectTag(rt.Field(i).Tag.Get("path"))
		} else {
			if !f.CanSet() { // unexported
				continue
			}

			s = s.Unmarshal(s.getSettableInterface(ft, f))
		}

		if s.Err != nil {
			break
		}
	}

	return s
}

func (s State) unmarshalArray(rt reflect.Type, rv reflect.Value) State {
	et := rt.Elem()
	fixed := hasFixedSize(et.Kind())

	disabled := s.DisableSep
	s.DisableSep = fixed

	n := rv.Len()
	for i := 0; i < n; i++ {
		s = s.Unmarshal(s.getSettableInterface(et, rv.Index(i)))
		if s.Err != nil {
			break
		}
	}

	s.DisableSep = disabled
	s.NeedSep = fixed

	return s.MaybeExpectSep()
}

func (s State) getSettableInterface(t reflect.Type, v reflect.Value) (i interface{}) {
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			nv := reflect.New(t.Elem())
			v.Set(nv)
			i = nv.Interface()
		} else {
			i = v.Interface()
		}
	} else {
		i = v.Addr().Interface()
	}

	return
}

func hasFixedSize(kind reflect.Kind) bool {
	return kind >= reflect.Bool && kind <= reflect.Complex128
}
