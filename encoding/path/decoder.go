package path

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"

	"github.com/iheartradio/cog"
)

// Unmarshaler is the interface implemented by objects that can unmarshal
// themselves from a []byte
type Unmarshaler interface {
	UnmarshalPath(d Decoder) Decoder
}

// A Decoder is used for Unmarshaling paths. Never create this directly; use
// NewDecoder instead.
type Decoder struct {
	State
}

// NewDecoder creates a new Decoder and checks that the first byte of the path
// is the appropriate separator.
func (s Separator) NewDecoder(b []byte) (dec Decoder) {
	if len(b) == 0 || b[0] != byte(s) {
		dec.Err = fmt.Errorf("path: first char of path is not %c", s)
	} else {
		dec.State = State{
			B: b[1:],
			s: byte(s),
		}
	}

	return
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func (s Separator) Unmarshal(b []byte, v interface{}) (unused []byte, err error) {
	d := s.NewDecoder(b)
	if d.Err != nil {
		unused = b
		err = d.Err
	} else {
		d = d.Unmarshal(v)
		unused = d.B
		err = d.Err
	}

	return
}

// MustUnmarshal is like Unmarshal, except it panics on failure
func (s Separator) MustUnmarshal(b []byte, p interface{}) (unused []byte) {
	unused, err := s.Unmarshal(b, p)
	cog.Must(err, "unmarshal failed")
	return unused
}

// Must ensures that there were no Unmarshaling errors, returning the unused
// portion of the path.
func (d Decoder) Must() []byte {
	cog.Must(d.Err, "unmarshal failed")
	return d.B
}

// Unmarshal unmarshals from the current state into the given value
func (d Decoder) Unmarshal(v interface{}) Decoder {
	switch v := v.(type) {
	case Unmarshaler:
		return v.UnmarshalPath(d)

	case *bool:
		return d.ExpectBool(v)

	case *int8:
		return d.ExpectInt8(v)

	case *int16:
		return d.ExpectInt16(v)

	case *int32:
		return d.ExpectInt32(v)

	case *int64:
		return d.ExpectInt64(v)

	case *uint8:
		return d.ExpectUint8(v)

	case *uint16:
		return d.ExpectUint16(v)

	case *uint32:
		return d.ExpectUint32(v)

	case *uint64:
		return d.ExpectUint64(v)

	case *float32:
		return d.ExpectFloat32(v)

	case *float64:
		return d.ExpectFloat64(v)

	case *complex64:
		return d.ExpectComplex64(v)

	case *complex128:
		return d.ExpectComplex128(v)

	case *string:
		return d.ExpectString(v)

	case *[]byte:
		return d.ExpectBytes(v)

	default:
		return d.unmarshalReflect(v)
	}
}

// ExpectSep looks for the delimiter as the next byte in the buffer
func (d Decoder) ExpectSep() Decoder {
	if len(d.B) == 0 || d.B[0] != d.s {
		d.Err = fmt.Errorf("path: invalid path: missing delimiter")
	} else {
		d.B = d.B[1:]
	}

	return d
}

// ExpectBool decodes an encoded boolean value from the buffer
func (d Decoder) ExpectBool(v *bool) Decoder {
	if len(d.B) == 0 {
		d.Err = fmt.Errorf("path: invalid path: bool truncated")
		return d
	}

	*v = d.B[0] != '\x00'
	d.B = d.B[1:]
	return d.ExpectSep()
}

func (d Decoder) readInt(size int) (int64, Decoder) {
	if len(d.B) < size {
		d.Err = fmt.Errorf("path: invalid path: need %d for int", size)
		return 0, d
	}

	ib := make([]byte, 8)
	copy(ib[8-size:], d.B)

	d.B = d.B[size:]
	v := (int64(ib[0]) << 56) |
		(int64(ib[1]) << 48) |
		(int64(ib[2]) << 40) |
		(int64(ib[3]) << 32) |
		(int64(ib[4]) << 24) |
		(int64(ib[5]) << 16) |
		(int64(ib[6]) << 8) |
		(int64(ib[7]))

	return v, d.ExpectSep()
}

func (d Decoder) readUint(size int) (uint64, Decoder) {
	if len(d.B) < size {
		d.Err = fmt.Errorf("path: invalid path: need %d for int", size)
		return 0, d
	}

	ib := make([]byte, 8)
	copy(ib[8-size:], d.B)

	d.B = d.B[size:]
	v := (uint64(ib[0]) << 56) |
		(uint64(ib[1]) << 48) |
		(uint64(ib[2]) << 40) |
		(uint64(ib[3]) << 32) |
		(uint64(ib[4]) << 24) |
		(uint64(ib[5]) << 16) |
		(uint64(ib[6]) << 8) |
		(uint64(ib[7]))

	return v, d.ExpectSep()
}

// ExpectInt8 decodes an encoded int8 value from the buffer
func (d Decoder) ExpectInt8(v *int8) Decoder {
	i, d := d.readInt(1)
	if d.Err == nil {
		*v = int8(i)
	}

	return d
}

// ExpectInt16 decodes an encoded int16 value from the buffer
func (d Decoder) ExpectInt16(v *int16) Decoder {
	i, d := d.readInt(2)
	if d.Err == nil {
		*v = int16(i)
	}

	return d
}

// ExpectInt32 decodes an encoded int32 value from the buffer
func (d Decoder) ExpectInt32(v *int32) Decoder {
	i, d := d.readInt(4)
	if d.Err == nil {
		*v = int32(i)
	}

	return d
}

// ExpectInt64 decodes an encoded int64 value from the buffer
func (d Decoder) ExpectInt64(v *int64) Decoder {
	i, d := d.readInt(8)
	if d.Err == nil {
		*v = int64(i)
	}

	return d
}

// ExpectUint8 decodes an encoded uint8 value from the buffer
func (d Decoder) ExpectUint8(v *uint8) Decoder {
	i, d := d.readUint(1)
	if d.Err == nil {
		*v = uint8(i)
	}
	return d
}

// ExpectUint16 decodes an encoded uint16 value from the buffer
func (d Decoder) ExpectUint16(v *uint16) Decoder {
	i, d := d.readUint(2)
	if d.Err == nil {
		*v = uint16(i)
	}
	return d
}

// ExpectUint32 decodes an encoded uint32 value from the buffer
func (d Decoder) ExpectUint32(v *uint32) Decoder {
	i, d := d.readUint(4)
	if d.Err == nil {
		*v = uint32(i)
	}
	return d
}

// ExpectUint64 decodes an encoded uint64 value from the buffer
func (d Decoder) ExpectUint64(v *uint64) Decoder {
	i, d := d.readUint(8)
	if d.Err == nil {
		*v = uint64(i)
	}
	return d
}

// ExpectFloat32 decodes an encoded uint64 value from the buffer
func (d Decoder) ExpectFloat32(v *float32) Decoder {
	i, d := d.readUint(4)
	if d.Err == nil {
		*v = math.Float32frombits(uint32(i))
	}
	return d
}

// ExpectFloat64 decodes an encoded uint64 value from the buffer
func (d Decoder) ExpectFloat64(v *float64) Decoder {
	i, d := d.readUint(8)
	if d.Err == nil {
		*v = math.Float64frombits(uint64(i))
	}
	return d
}

// ExpectComplex64 decodes an encoded complex64 value from the buffer
func (d Decoder) ExpectComplex64(v *complex64) Decoder {
	i, d := d.readUint(8)
	if d.Err == nil {
		rl := math.Float32frombits(uint32(i >> 32))
		ig := math.Float32frombits(uint32(i))
		*v = complex(rl, ig)
	}
	return d
}

// ExpectComplex128 decodes an encoded complex128 value from the buffer
func (d Decoder) ExpectComplex128(v *complex128) Decoder {
	if d.Err == nil && len(d.B) < 16 {
		d.Err = fmt.Errorf("path: invalid path: complex128 truncated")
	}

	if d.Err != nil {
		return d
	}

	rl := math.Float64frombits(binary.BigEndian.Uint64(d.B))
	ig := math.Float64frombits(binary.BigEndian.Uint64(d.B[8:]))
	*v = complex(rl, ig)

	d.B = d.B[16:]

	return d.ExpectSep()
}

// ExpectString decodes an encoded string value from the buffer
func (d Decoder) ExpectString(v *string) Decoder {
	i := bytes.IndexByte(d.B, d.s)
	if i == -1 {
		d.Err = fmt.Errorf("path: failed to read string: missing separator")
		return d
	}

	*v = string(d.B[:i])
	d.B = d.B[i+1:]

	return d
}

// ExpectTag reads expects the next string it reads from the buffer to be the
// given tag
func (d Decoder) ExpectTag(tag string) Decoder {
	n := len(tag)
	if len(d.B) < n {
		d.Err = fmt.Errorf("path: failed to read tag: tag truncated")
		return d
	}

	got := string(d.B[:n])
	if got != tag {
		d.Err = fmt.Errorf("path: tag mismatch: %s != %s", got, tag)
		return d
	}

	d.B = d.B[n:]

	return d.ExpectSep()
}

// ExpectTagBytes reads expects the next byte slice it reads from the buffer to
// be the given tag. This is a faster version of ExpectTag since it requires no
// memory allocations (assuming you pass in a pre-allocated, read-only byte
// slice).
func (d Decoder) ExpectTagBytes(tag []byte) Decoder {
	n := len(tag)
	if len(d.B) < n {
		d.Err = fmt.Errorf("path: failed to read tag: tag truncated")
		return d
	}

	if !bytes.Equal(d.B[:n], tag) {
		d.Err = fmt.Errorf("path: tag mismatch: %s != %s",
			string(d.B[:n]),
			string(tag))
		return d
	}

	d.B = d.B[n:]

	return d.ExpectSep()
}

// ExpectBytes decodes an encoded byte slice value from the buffer
func (d Decoder) ExpectBytes(v *[]byte) Decoder {
	i := bytes.IndexByte(d.B, d.s)
	if i == -1 {
		d.Err = fmt.Errorf("path: failed to read []byte: missing separator")
		return d
	}

	*v = d.B[:i]
	d.B = d.B[i:]

	return d.ExpectSep()
}

// ExpectByteArray decodes an encoded byte slice value from the buffer
func (d Decoder) ExpectByteArray(v []byte) Decoder {
	n := copy(v, d.B)
	if n != len(v) {
		d.Err = fmt.Errorf("path: bytes truncated in array")
		return d
	}

	d.B = d.B[n:]

	return d.ExpectSep()
}

func (d Decoder) unmarshalReflect(v interface{}) Decoder {
	rt := reflect.TypeOf(v)

	// A nil obviously can't be a path
	if rt == nil {
		d.Err = fmt.Errorf("path: cannot unmarshal into a nil value")
		return d
	}

	kind := rt.Kind()
	if kind != reflect.Ptr {
		d.Err = fmt.Errorf("path: need a pointer to unmarshal into")
		return d
	}

	rv := reflect.Indirect(reflect.ValueOf(v))
	rt = rv.Type()
	kind = rt.Kind()

	switch kind {
	case reflect.Struct:
		return d.unmarshalStruct(rt, rv)

	case reflect.Array:
		return d.unmarshalArray(rt, rv)

	case reflect.Bool:
		return d.Unmarshal((*bool)(unsafe.Pointer(rv.UnsafeAddr())))

	case reflect.Int8:
		return d.Unmarshal((*int8)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Int16:
		return d.Unmarshal((*int16)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Int32:
		return d.Unmarshal((*int32)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Int64:
		return d.Unmarshal((*int64)(unsafe.Pointer(rv.UnsafeAddr())))

	case reflect.Uint8:
		return d.Unmarshal((*uint8)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Uint16:
		return d.Unmarshal((*uint16)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Uint32:
		return d.Unmarshal((*uint32)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Uint64:
		return d.Unmarshal((*uint64)(unsafe.Pointer(rv.UnsafeAddr())))

	case reflect.Float32:
		return d.Unmarshal((*float32)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Float64:
		return d.Unmarshal((*float64)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Complex64:
		return d.Unmarshal((*complex64)(unsafe.Pointer(rv.UnsafeAddr())))
	case reflect.Complex128:
		return d.Unmarshal((*complex128)(unsafe.Pointer(rv.UnsafeAddr())))

	case reflect.Ptr:
		d.Err = fmt.Errorf("path: sorry, pointers are not allowed")
		return d

	default:
		d.Err = fmt.Errorf("path: unsupported unmarshal kind: %s", kind)
		return d
	}
}

func (d Decoder) unmarshalStruct(rt reflect.Type, v reflect.Value) Decoder {
	n := v.NumField()
	for i := 0; i < n; i++ {
		f := v.Field(i)

		if f.Type() == staticType {
			d = d.ExpectTag(rt.Field(i).Tag.Get("path"))
		} else {
			if !f.CanSet() { // unexported
				continue
			}

			d = d.Unmarshal(f.Addr().Interface())
		}

		if d.Err != nil {
			break
		}
	}

	return d
}

func (d Decoder) unmarshalArray(rt reflect.Type, rv reflect.Value) Decoder {
	n := rv.Len()
	for i := 0; i < n; i++ {
		d = d.Unmarshal(rv.Index(i).Addr().Interface())
		if d.Err != nil {
			break
		}
	}

	return d
}
