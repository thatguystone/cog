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
	"io"
	"reflect"
	"strings"

	"github.com/thatguystone/cog"
)

// Separator allows you to change the path separator used.
type Separator struct {
	b  byte
	s  string
	bs []byte
}

// Static is used to place a static, unchanging element into the path.
type Static struct{}

// Marshaler is the interface implemented by objects that can marshal themselves
// into a valid path
type Marshaler interface {
	MarshalPath(buff *bytes.Buffer, s Separator) error
}

// Unmarshaler is the interface implemented by objects that can unmarshal
// themselves from a bytes.Buffer.
type Unmarshaler interface {
	UnmarshalPath(buff *bytes.Buffer, s Separator) error
}

var staticType = reflect.TypeOf(Static{})

// NewSeparator uses the given delimiter instead of "/".
func NewSeparator(delim byte) Separator {
	return Separator{
		b:  delim,
		s:  string(delim),
		bs: []byte{delim},
	}
}

// Marshal turns the given struct into a structured path
func (s Separator) Marshal(v interface{}) (b []byte, err error) {
	buff := bytes.Buffer{}
	err = s.MarshalInto(v, &buff)
	if err == nil {
		b = buff.Bytes()
	}

	return
}

// MustMarshal is like Marshal, except it panics on failure
func (s Separator) MustMarshal(v interface{}) []byte {
	b, err := s.Marshal(v)
	cog.Must(err, "marshal failed")
	return b
}

// MarshalInto works exactly like Marshal, except it writes the path to the
// given Buffer instead of returning a [ ]byte.
func (s Separator) MarshalInto(v interface{}, buff *bytes.Buffer) error {
	err := s.marshalInto(v, buff, true)
	if err == nil && buff.Len() > 1 {
		s.EmitDelim(buff)
	}

	return err
}

// MustMarshalInto is like MarshalInto, except it panics on failure
func (s Separator) MustMarshalInto(v interface{}, buff *bytes.Buffer) {
	err := s.MarshalInto(v, buff)
	cog.Must(err, "marshal failed")
}

func (s Separator) marshalInto(v interface{}, buff *bytes.Buffer, needDelim bool) (err error) {
	if needDelim {
		s.EmitDelim(buff)
	}

	switch v := v.(type) {
	case Marshaler:
		return v.MarshalPath(buff, s)

	case bool:
		s.EmitBool(v, buff)
	case *bool:
		s.EmitBool(*v, buff)

	case int8:
		s.EmitInt8(v, buff)
	case *int8:
		s.EmitInt8(*v, buff)

	case int16:
		s.EmitInt16(v, buff)
	case *int16:
		s.EmitInt16(*v, buff)

	case int32:
		s.EmitInt32(v, buff)
	case *int32:
		s.EmitInt32(*v, buff)

	case int64:
		s.EmitInt64(v, buff)
	case *int64:
		s.EmitInt64(*v, buff)

	case uint8:
		s.EmitUint8(v, buff)
	case *uint8:
		s.EmitUint8(*v, buff)

	case uint16:
		s.EmitUint16(v, buff)
	case *uint16:
		s.EmitUint16(*v, buff)

	case uint32:
		s.EmitUint32(v, buff)
	case *uint32:
		s.EmitUint32(*v, buff)

	case uint64:
		s.EmitUint64(v, buff)
	case *uint64:
		s.EmitUint64(*v, buff)

	case float32, float64, complex64, complex128,
		*float32, *float64, *complex64, *complex128:
		err = binary.Write(buff, binary.BigEndian, v)

	case string:
		return s.EmitString(v, buff)
	case *string:
		return s.EmitString(*v, buff)

	case []byte:
		return s.EmitBytes(v, buff)
	case *[]byte:
		return s.EmitBytes(*v, buff)

	default:
		return s.marshalReflect(reflect.ValueOf(v), buff, false)
	}

	return
}

// EmitDelim writes the path separator into the buffer
func (s Separator) EmitDelim(buff *bytes.Buffer) {
	buff.WriteByte(s.b)
}

// EmitBool writes an encoded boolean value into the buffer
func (Separator) EmitBool(v bool, buff *bytes.Buffer) {
	b := byte('\x00')
	if v {
		b = '\x01'
	}

	buff.WriteByte(b)
}

// EmitInt writes an encoded int value into the buffer
func (Separator) EmitInt(v int64, buff *bytes.Buffer, size int) {
	sign := byte(0)
	if v < 0 {
		v = -v
		sign = 0x80
	}

	ib := make([]byte, 8)
	binary.BigEndian.PutUint64(ib, uint64(v))

	ibs := ib[8-size:]
	ibs[0] |= sign

	buff.Write(ibs)
}

// EmitUint writes an encoded uint value into the buffer
func (Separator) EmitUint(v uint64, buff *bytes.Buffer, size int) {
	ib := make([]byte, 8)
	binary.BigEndian.PutUint64(ib, v)
	buff.Write(ib[8-size:])
}

// EmitInt8 is a wrapper around EmitIn that writes an int8
func (s Separator) EmitInt8(v int8, buff *bytes.Buffer) {
	s.EmitInt(int64(v), buff, 1)
}

// EmitInt16 is a wrapper around EmitInt that writes an int16
func (s Separator) EmitInt16(v int16, buff *bytes.Buffer) {
	s.EmitInt(int64(v), buff, 2)
}

// EmitInt32 is a wrapper around EmitInt that writes an int32
func (s Separator) EmitInt32(v int32, buff *bytes.Buffer) {
	s.EmitInt(int64(v), buff, 4)
}

// EmitInt64 is a wrapper around EmitInt that writes an int64
func (s Separator) EmitInt64(v int64, buff *bytes.Buffer) {
	s.EmitInt(v, buff, 8)
}

// EmitUint8 is a wrapper around EmitUin that writes an uint8
func (s Separator) EmitUint8(v uint8, buff *bytes.Buffer) {
	s.EmitUint(uint64(v), buff, 1)
}

// EmitUint16 is a wrapper around EmitUint that writes an uint16
func (s Separator) EmitUint16(v uint16, buff *bytes.Buffer) {
	s.EmitUint(uint64(v), buff, 2)
}

// EmitUint32 is a wrapper around EmitUint that writes an uint32
func (s Separator) EmitUint32(v uint32, buff *bytes.Buffer) {
	s.EmitUint(uint64(v), buff, 4)
}

// EmitUint64 is a wrapper around EmitUint that writes an uint64
func (s Separator) EmitUint64(v uint64, buff *bytes.Buffer) {
	s.EmitUint(v, buff, 8)
}

// EmitString writes an encoded string value into the buffer
func (s Separator) EmitString(v string, buff *bytes.Buffer) error {
	if strings.IndexByte(v, s.b) != -1 {
		return fmt.Errorf("path: invalid string: may not contain a `%s`", s.s)
	}

	_, err := buff.WriteString(v)
	return err
}

// EmitBytes writes an encoded byte slice value into the buffer
func (s Separator) EmitBytes(v []byte, buff *bytes.Buffer) error {
	if bytes.IndexByte(v, s.b) != -1 {
		return fmt.Errorf("path: invalid []byte: may not contain a `%s`", s.s)
	}

	_, err := buff.Write(v)
	return err
}

func (s Separator) marshalReflect(
	rv reflect.Value,
	buff *bytes.Buffer,
	needDelim bool) error {

	var rt reflect.Type
	if rv.IsValid() {
		rt = rv.Type()
	}

	// A nil obviously can't be a path
	if rt == nil {
		return fmt.Errorf("path: a nil cannot be turned into a path")
	}

	kind := rt.Kind()
	if kind == reflect.Ptr {
		return s.marshalReflect(reflect.Indirect(rv), buff, needDelim)
	}

	switch kind {
	case reflect.Struct:
		return s.marshalStruct(rt, rv, buff, needDelim)

	case reflect.Array:
		return s.marshalArray(rv, buff, needDelim)

	default:
		return fmt.Errorf("path: unsupported marshal kind: %s", kind)
	}
}

func (s Separator) marshalStruct(
	rt reflect.Type,
	rv reflect.Value,
	buff *bytes.Buffer,
	needDelim bool) (err error) {

	n := rv.NumField()
	for i := 0; i < n; i++ {
		f := rv.Field(i)

		if f.Type() == staticType {
			err = s.marshalInto(rt.Field(i).Tag.Get("path"), buff, needDelim)
		} else {
			if !f.CanInterface() { // unexported
				continue
			}

			if f.Kind() == reflect.Ptr {
				if f.IsNil() {
					return fmt.Errorf("path: cannot Marshal into a nil pointer")
				}

				f = reflect.Indirect(f)
			}

			err = s.marshalInto(f.Interface(), buff, needDelim)
		}

		if err != nil {
			return
		}

		needDelim = true
	}

	return
}

func (s Separator) marshalArray(
	rv reflect.Value,
	buff *bytes.Buffer,
	needDelim bool) (err error) {

	n := rv.Len()
	for i := 0; i < n; i++ {
		err = s.marshalInto(rv.Index(i).Interface(), buff, needDelim)
		if err != nil {
			break
		}

		needDelim = true
	}

	return
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func (s Separator) Unmarshal(b []byte, v interface{}) error {
	buff := bytes.NewBuffer(b)
	return s.UnmarshalFrom(buff, v)
}

// MustUnmarshal is like Unmarshal, except it panics on failure
func (s Separator) MustUnmarshal(b []byte, p interface{}) {
	cog.Must(Unmarshal(b, p), "unmarshal failed")
}

// UnmarshalFrom works exactly like Unmarshal, except it reads from the given
// Buffer instead of a [ ]byte.
func (s Separator) UnmarshalFrom(buff *bytes.Buffer, v interface{}) error {
	return s.unmarshalFrom(buff, v, true)
}

// MustUnmarshalFrom is like UnmarshalFrom, except it panics on failure
func (s Separator) MustUnmarshalFrom(buff *bytes.Buffer, p interface{}) {
	cog.Must(UnmarshalFrom(buff, p), "unmarshal failed")
}

func (s Separator) unmarshalFrom(
	buff *bytes.Buffer,
	v interface{},
	expectDelim bool) error {

	if expectDelim {
		err := s.ExpectDelim(buff)
		if err != nil {
			return err
		}
	}

	switch v := v.(type) {
	case Unmarshaler:
		return v.UnmarshalPath(buff, s)

	case *bool:
		return s.ExpectBool(v, buff)

	case *int8:
		return s.ExpectInt8(v, buff)

	case *int16:
		return s.ExpectInt16(v, buff)

	case *int32:
		return s.ExpectInt32(v, buff)

	case *int64:
		return s.ExpectInt64(v, buff)

	case *uint8:
		return s.ExpectUint8(v, buff)

	case *uint16:
		return s.ExpectUint16(v, buff)

	case *uint32:
		return s.ExpectUint32(v, buff)

	case *uint64:
		return s.ExpectUint64(v, buff)

	case *float32, *float64, *complex64, *complex128:
		return binary.Read(buff, binary.BigEndian, v)

	case *string:
		return s.ExpectString(v, buff)

	case *[]byte:
		return s.ExpectBytes(v, buff)

	default:
		return s.unmarshalReflect(v, buff, false)
	}
}

// ExpectDelim looks for the delimiter as the next byte in the buffer
func (s Separator) ExpectDelim(buff *bytes.Buffer) error {
	c, err := buff.ReadByte()
	if c != s.b || err != nil {
		err = fmt.Errorf("path: invalid path: missing delimiter")
	}

	return err
}

// ExpectBool decodes an encoded boolean value from the buffer
func (Separator) ExpectBool(v *bool, buff *bytes.Buffer) (err error) {
	var c byte
	c, err = buff.ReadByte()
	if err == nil {
		*v = c != '\x00'
	}

	return
}

// ExpectInt decodes an encoded int value from the buffer
func (Separator) ExpectInt(buff *bytes.Buffer, size int) (int64, error) {
	ib := make([]byte, 8)
	ibs := ib[8-size:]

	n, err := buff.Read(ibs)
	if err != nil || n != len(ibs) {
		return 0, fmt.Errorf("path: failed to read int")
	}

	sign := ibs[0]&0x80 != 0
	if sign {
		ibs[0] ^= 0x80

		min := true
		for _, b := range ibs {
			if b != 0 {
				min = false
				break
			}
		}

		if min {
			ibs[0] |= 0x80
		}
	}

	v := int64(binary.BigEndian.Uint64(ib))

	if sign {
		v = -v
	}

	return v, nil
}

// ExpectUint decodes an encoded uint value from the buffer
func (Separator) ExpectUint(buff *bytes.Buffer, size int) (uint64, error) {
	ib := make([]byte, 8)
	ibs := ib[8-size:]

	n, err := buff.Read(ibs)
	if err != nil || n != len(ibs) {
		return 0, fmt.Errorf("path: failed to read uint")
	}

	return binary.BigEndian.Uint64(ib), nil
}

// ExpectInt8 decodes an encoded int8 value from the buffer
func (s Separator) ExpectInt8(v *int8, buff *bytes.Buffer) error {
	i, err := s.ExpectInt(buff, 1)
	if err == nil {
		*v = int8(i)
	}
	return err
}

// ExpectInt16 decodes an encoded int16 value from the buffer
func (s Separator) ExpectInt16(v *int16, buff *bytes.Buffer) error {
	i, err := s.ExpectInt(buff, 2)
	if err == nil {
		*v = int16(i)
	}
	return err
}

// ExpectInt32 decodes an encoded int32 value from the buffer
func (s Separator) ExpectInt32(v *int32, buff *bytes.Buffer) error {
	i, err := s.ExpectInt(buff, 4)
	if err == nil {
		*v = int32(i)
	}
	return err
}

// ExpectInt64 decodes an encoded int64 value from the buffer
func (s Separator) ExpectInt64(v *int64, buff *bytes.Buffer) error {
	i, err := s.ExpectInt(buff, 8)
	if err == nil {
		*v = int64(i)
	}
	return err
}

// ExpectUint8 decodes an encoded uint8 value from the buffer
func (s Separator) ExpectUint8(v *uint8, buff *bytes.Buffer) error {
	i, err := s.ExpectUint(buff, 1)
	if err == nil {
		*v = uint8(i)
	}
	return err
}

// ExpectUint16 decodes an encoded uint16 value from the buffer
func (s Separator) ExpectUint16(v *uint16, buff *bytes.Buffer) error {
	i, err := s.ExpectUint(buff, 2)
	if err == nil {
		*v = uint16(i)
	}
	return err
}

// ExpectUint32 decodes an encoded uint32 value from the buffer
func (s Separator) ExpectUint32(v *uint32, buff *bytes.Buffer) error {
	i, err := s.ExpectUint(buff, 4)
	if err == nil {
		*v = uint32(i)
	}
	return err
}

// ExpectUint64 decodes an encoded uint64 value from the buffer
func (s Separator) ExpectUint64(v *uint64, buff *bytes.Buffer) error {
	i, err := s.ExpectUint(buff, 8)
	if err == nil {
		*v = uint64(i)
	}
	return err
}

// ExpectString decodes an encoded string value from the buffer
func (s Separator) ExpectString(v *string, buff *bytes.Buffer) (err error) {
	ss, err := buff.ReadString(s.b)
	if err == nil {
		// Put the delim back
		buff.UnreadByte()

		*v = ss[:len(ss)-1]
	}

	return
}

// ExpectTag reads expects the next string it reads from the buffer to be the
// given tag
func (s Separator) ExpectTag(tag string, buff *bytes.Buffer) (err error) {
	ss, err := buff.ReadString(s.b)
	if err == nil {
		// Put the delim back
		buff.UnreadByte()

		ss = ss[:len(ss)-1]
		if ss != tag {
			err = fmt.Errorf("path: tag mismatch: %s != %s", ss, tag)
		}
	}

	return
}

// ExpectTagBytes reads expects the next byte slice it reads from the buffer to
// be the given tag. This is a faster version of ExpectTag since it requires no
// memory allocations (assuming you pass in a pre-allocated, read-only byte
// slice).
func (s Separator) ExpectTagBytes(tag []byte, buff *bytes.Buffer) (err error) {
	i := bytes.IndexByte(buff.Bytes(), s.b)
	if i == -1 {
		err = io.EOF
	} else {
		b := buff.Next(i)
		if !bytes.Equal(b, tag) {
			err = fmt.Errorf("path: tag mismatch: %s != %s",
				string(b),
				string(tag))
		}
	}

	return
}

// ExpectBytes decodes an encoded byte slice value from the buffer
func (s Separator) ExpectBytes(v *[]byte, buff *bytes.Buffer) (err error) {
	b, err := buff.ReadBytes(s.b)
	if err == nil {
		// Put the delim back
		buff.UnreadByte()

		*v = b[:len(b)-1]
	}

	return
}

func (s Separator) unmarshalReflect(
	v interface{},
	buff *bytes.Buffer,
	expectDelim bool) error {

	rt := reflect.TypeOf(v)

	// A nil obviously can't be a path
	if rt == nil {
		return fmt.Errorf("path: cannot unmarshal into a nil value")
	}

	kind := rt.Kind()
	if kind != reflect.Ptr {
		return fmt.Errorf("path: need a pointer to unmarshal into")
	}

	rv := reflect.Indirect(reflect.ValueOf(v))
	rt = rv.Type()
	kind = rt.Kind()

	switch kind {
	case reflect.Struct:
		return s.unmarshalStruct(rt, rv, buff, expectDelim)

	case reflect.Array:
		return s.unmarshalArray(rv, buff, expectDelim)

	default:
		return fmt.Errorf("path: unsupported unmarshal kind: %s", kind)
	}
}

func (s Separator) unmarshalStruct(
	rt reflect.Type,
	v reflect.Value,
	buff *bytes.Buffer,
	expectDelim bool) (err error) {

	n := v.NumField()
	for i := 0; i < n; i++ {
		f := v.Field(i)

		if f.Type() == staticType {
			var ss string
			err = s.unmarshalFrom(
				buff,
				&ss,
				expectDelim)
			tag := rt.Field(i).Tag.Get("path")
			if err == nil && ss != tag {
				err = fmt.Errorf("path: tag mismatch: %s != %s", ss, tag)
			}
		} else {
			if !f.CanSet() { // unexported
				continue
			}

			var i interface{}

			if f.Kind() == reflect.Ptr {
				i = f.Interface()
			} else {
				i = f.Addr().Interface()
			}

			err = s.unmarshalFrom(buff, i, expectDelim)
		}

		if err != nil {
			return
		}

		expectDelim = true
	}

	return
}

func (s Separator) unmarshalArray(
	rv reflect.Value,
	buff *bytes.Buffer,
	needDelim bool) (err error) {

	n := rv.Len()
	for i := 0; i < n; i++ {
		err = s.unmarshalFrom(
			buff,
			rv.Index(i).Addr().Interface(),
			needDelim)
		if err != nil {
			break
		}

		needDelim = true
	}

	return
}
