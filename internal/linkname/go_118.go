//go:build go1.18
// +build go1.18

package linkname

import (
	"reflect"
	"unsafe"
)

func init() {
	mapIterTyp := reflect.TypeOf(reflect.MapIter{})
	hiterField, ok := mapIterTyp.FieldByName("hiter")
	if !ok {
		panic("reflect.MapIter field iter not found")
	}
	hiterType = toRType(hiterField.Type)
}

var hiterType unsafe.Pointer // *reflect.rtype

func Reflect_mapiterinit(rtype unsafe.Pointer, m unsafe.Pointer) unsafe.Pointer {
	hiter := Reflect_unsafe_New(hiterType)
	reflect_mapiterinit(rtype, m, hiter)
	return hiter
}

// reflect_mapiterinit .
// m escapes into the return value, but the caller of Reflect_mapiterinit
// doesn't let the return value escape.
//go:noescape
//go:linkname reflect_mapiterinit reflect.mapiterinit
func reflect_mapiterinit(rtype unsafe.Pointer, m unsafe.Pointer, hiter unsafe.Pointer)
