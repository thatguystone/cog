//go:build appengine || purego

package check

//gocovr:skip-file

import "reflect"

func forceCanInterface(rv reflect.Value) (reflect.Value, bool) {
	return reflect.Value{}, false
}
