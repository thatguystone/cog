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
	"reflect"
	"strings"
)

// Separator allows you to change the path separator used.
type Separator struct {
	b  byte
	s  string
	bs []byte
}

// Static is used to place a static, unchanging element into the path.
type Static struct{}

var staticType = reflect.TypeOf(Static{})

// NewSeparator is used to change the path separator
func NewSeparator(delim byte) Separator {
	return Separator{
		b:  delim,
		s:  string(delim),
		bs: []byte{delim},
	}
}

// Marshal turns the given struct into a structured path
func (s Separator) Marshal(p interface{}) (b []byte, err error) {
	buff := bytes.Buffer{}
	err = s.MarshalInto(p, &buff)
	if err == nil {
		b = buff.Bytes()
	}

	return
}

// MarshalInto works exactly like Marshal, except it writes the path to the
// given Buffer instead of returning a [ ]byte.
func (s Separator) MarshalInto(p interface{}, buff *bytes.Buffer) error {
	err := s.marshalInto(p, buff, true)
	if err == nil {
		buff.WriteByte(s.b)
	}

	return err
}

func (s Separator) marshalInto(p interface{}, buff *bytes.Buffer, needDelim bool) (err error) {
	pt := reflect.TypeOf(p)

	// A nil obviously returns a nil path
	if pt == nil {
		return fmt.Errorf("path: a nil cannot be turned into a path")
	}

	pv := reflect.ValueOf(p)

	if pt.Kind() == reflect.Ptr {
		if pv.IsNil() {
			return fmt.Errorf("path: cannot Marshal a nil pointer, expected a struct")
		}

		pv = reflect.Indirect(pv)
		pt = pv.Type()
	}

	if pt.Kind() != reflect.Struct {
		return fmt.Errorf("path: paths may only be read into structs")
	}

	n := pt.NumField()
	for i := 0; i < n; i++ {
		v := pv.Field(i)

		if !v.CanInterface() { // unexported
			continue
		}

		if needDelim {
			buff.WriteByte(s.b)
		}

		needDelim = true

		switch v.Kind() {
		case reflect.Bool:
			b := byte('\x00')
			if v.Bool() {
				b = '\x01'
			}

			buff.WriteByte(b)

		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i := v.Int()
			sign := byte(0)
			if i < 0 {
				i = -i
				sign = 0x80
			}

			ib := make([]byte, 8)
			binary.BigEndian.PutUint64(ib, uint64(i))
			ibs := ib[8-v.Type().Size():]
			ibs[0] |= sign

			_, err = buff.Write(ibs)

		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ib := make([]byte, 8)
			binary.BigEndian.PutUint64(ib, v.Uint())
			_, err = buff.Write(ib[8-v.Type().Size():])

		case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
			err = binary.Write(buff, binary.BigEndian, v.Interface())

		case reflect.String:
			err = s.writeString(buff, v.String())

		case reflect.Struct:
			if v.Type() == staticType {
				err = s.writeString(buff, pt.Field(i).Tag.Get("path"))
			} else {
				err = s.marshalInto(v.Interface(), buff, false)
			}

		default:
			err = fmt.Errorf("path: unsupported type: %s", v.Kind())
		}

		if err != nil {
			return
		}
	}

	return
}

func (s Separator) writeString(buff *bytes.Buffer, ss string) error {
	if strings.Contains(ss, s.s) {
		return fmt.Errorf("path: invalid string: may not contain a `%s`", s.s)
	}

	_, err := buff.WriteString(ss)
	return err
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func (s Separator) Unmarshal(b []byte, p interface{}) error {
	buff := bytes.NewBuffer(b)
	return s.UnmarshalFrom(buff, p)
}

// UnmarshalFrom works exactly like Unmarshal, except it reads from the given
// Buffer instead of a [ ]byte.
func (s Separator) UnmarshalFrom(buff *bytes.Buffer, p interface{}) (err error) {
	return s.unmarshalFrom(buff, p, true)
}

func (s Separator) unmarshalFrom(buff *bytes.Buffer, p interface{}, expectDelim bool) (err error) {
	pv := reflect.ValueOf(p)

	if pv.Kind() != reflect.Ptr || pv.IsNil() {
		return fmt.Errorf("path: expected pointer to unmarshal into")
	}

	pv = reflect.Indirect(pv)
	pt := pv.Type()

	readString := func() (val string, err error) {
		expectDelim = false

		val, err = buff.ReadString(s.b)
		if err == nil {
			val = val[:len(val)-1]
		}

		return
	}

	n := pv.NumField()
	for i := 0; i < n; i++ {
		v := pv.Field(i)

		if !v.CanSet() {
			continue
		}

		if expectDelim {
			c, err := buff.ReadByte()
			if c != s.b || err != nil {
				return fmt.Errorf("path: invalid path: missing delimiter")
			}
		}

		expectDelim = true

		switch v.Kind() {
		case reflect.Bool:
			var c byte
			c, err = buff.ReadByte()
			if err == nil {
				v.SetBool(c != '\x00')
			}

		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var ibs []byte
			ib := make([]byte, 8)
			ibs, err = readInt(ib, buff, v)
			if err == nil {
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

				val := int64(binary.BigEndian.Uint64(ib))

				if sign {
					val = -val
				}

				v.SetInt(val)
			}

		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ib := make([]byte, 8)
			_, err = readInt(ib, buff, v)
			if err == nil {
				v.SetUint(binary.BigEndian.Uint64(ib))
			}

		case reflect.Float32:
			val := float32(0)
			err = binary.Read(buff, binary.BigEndian, &val)
			if err == nil {
				v.SetFloat(float64(val))
			}

		case reflect.Float64:
			val := float64(0)
			err = binary.Read(buff, binary.BigEndian, &val)
			if err == nil {
				v.SetFloat(val)
			}

		case reflect.Complex64:
			val := complex64(0)
			err = binary.Read(buff, binary.BigEndian, &val)
			if err == nil {
				v.SetComplex(complex128(val))
			}

		case reflect.Complex128:
			val := complex128(0)
			err = binary.Read(buff, binary.BigEndian, &val)
			if err == nil {
				v.SetComplex(val)
			}

		case reflect.String:
			var val string
			val, err = readString()
			if err == nil {
				v.SetString(val)
			}

		case reflect.Struct:
			if v.Type() == staticType {
				var val string
				val, err = readString()
				tag := pt.Field(i).Tag.Get("path")
				if err == nil && val != tag {
					err = fmt.Errorf("path: tag mismatch: %s != %s", val, tag)
				}
			} else {
				err = s.unmarshalFrom(buff, v.Addr().Interface(), false)
			}

		default:
			err = fmt.Errorf("path: unsupported type: %s", v.Kind())
		}

		if err != nil {
			return
		}
	}

	return
}

func readInt(ib []byte, buff *bytes.Buffer, v reflect.Value) ([]byte, error) {
	size := v.Type().Size()
	ibs := ib[8-size:]

	n, err := buff.Read(ibs)
	if err != nil || n != len(ibs) {
		return nil, fmt.Errorf("path: failed to read %s", v.Kind())
	}

	return ibs, nil
}
