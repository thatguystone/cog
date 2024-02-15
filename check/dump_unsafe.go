//go:build !(appengine || purego)

package check

//gocovr:skip-file

import "reflect"

var rvPtrField int

func init() {
	ft, _ := reflect.TypeOf(reflect.Value{}).FieldByName("ptr")
	if len(ft.Index) != 1 {
		panic(ft.Index)
	}
	rvPtrField = ft.Index[0]
}

func forceCanInterface(rv reflect.Value) (reflect.Value, bool) {
	uptr := reflect.ValueOf(rv).Field(rvPtrField).UnsafePointer()
	return reflect.NewAt(rv.Type(), uptr).Elem(), true
}
