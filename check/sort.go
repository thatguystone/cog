package check

import (
	"cmp"
	"fmt"
	"reflect"
	"sort"
)

type kv struct {
	k reflect.Value
	v reflect.Value
}

func sortMap(rv reflect.Value) []kv {
	kvs := make([]kv, 0, rv.Len())
	for iter := rv.MapRange(); iter.Next(); {
		kvs = append(kvs, kv{
			k: iter.Key(),
			v: iter.Value(),
		})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return compare(kvs[i].k, kvs[j].k) < 0
	})

	return kvs
}

func compare(av, bv reflect.Value) int {
	if c := compareTypes(av.Type(), bv.Type()); c != 0 {
		return c
	}

	switch av.Kind() {
	case reflect.Bool:
		return compareBool(av.Bool(), bv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(av.Int(), bv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return cmp.Compare(av.Uint(), bv.Uint())
	case reflect.Float32, reflect.Float64:
		return cmp.Compare(av.Float(), bv.Float())
	case reflect.Complex64, reflect.Complex128:
		return compareComplex(av.Complex(), bv.Complex())
	case reflect.String:
		return cmp.Compare(av.String(), bv.String())
	case reflect.Interface, reflect.Pointer:
		if c, ok := compareNil(av, bv); ok {
			return c
		}
		return compare(av.Elem(), bv.Elem())
	case reflect.Chan, reflect.UnsafePointer:
		if c, ok := compareNil(av, bv); ok {
			return c
		}
		return cmp.Compare(av.Pointer(), bv.Pointer())
	case reflect.Struct:
		for i := 0; i < av.NumField(); i++ {
			if c := compare(av.Field(i), bv.Field(i)); c != 0 {
				return c
			}
		}
		return 0
	case reflect.Array:
		for i := 0; i < av.Len(); i++ {
			if c := compare(av.Index(i), bv.Index(i)); c != 0 {
				return c
			}
		}
		return 0
	default:
		// Certain types cannot appear as keys (maps, funcs, slices), but be explicit.
		panic(fmt.Errorf("bad type in compare: %s", av.Type()))
	}
}

func compareTypes(at, bt reflect.Type) int {
	if at == bt {
		return 0
	}

	ak := at.Kind()
	bk := bt.Kind()

	if c := cmp.Compare(ak, bk); c != 0 {
		return c
	}

	if at.String() == ak.String() {
		return -1
	}

	if bt.String() == bk.String() {
		return +1
	}

	switch ak {
	case reflect.Pointer, reflect.Array, reflect.Chan:
		return compareTypes(at.Elem(), bt.Elem())
	default:
		return cmp.Compare(at.String(), bt.String())
	}
}

func compareBool(a, b bool) int {
	if a == b {
		return 0
	}

	if a {
		return +1
	}

	return -1
}

func compareComplex(a, b complex128) int {
	c := cmp.Compare(real(a), real(b))
	if c != 0 {
		return c
	}

	return cmp.Compare(imag(a), imag(b))
}

func compareNil(av, bv reflect.Value) (int, bool) {
	if av.IsNil() {
		if bv.IsNil() {
			return 0, true
		}
		return -1, true
	}
	if bv.IsNil() {
		return +1, true
	}
	return 0, false
}
