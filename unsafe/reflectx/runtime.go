package reflectx

import (
	"reflect"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

// ArrayAt returns the i-th element of p,
// an array whose elements are elemSize bytes wide.
// The array pointed at by p must have at least i+1 elements:
// it is invalid (but impossible to check here) to pass i >= len,
// because then the result will point outside the array.
func ArrayAt(p unsafe.Pointer, i int, elemSize uintptr) unsafe.Pointer {
	return unsafe.Add(p, uintptr(i)*elemSize)
}

// MakeSlice makes a new slice of the given reflect.Type and length, capacity.
func MakeSlice(elemTyp reflect.Type, length, capacity int) (slice any, header *SliceHeader) {
	elemRType := ToRType(elemTyp)
	data := linkname.Reflect_unsafe_NewArray(unsafe.Pointer(elemRType), capacity)
	header = &SliceHeader{
		Data: data,
		Len:  length,
		Cap:  capacity,
	}
	slice = SliceOf(elemRType).PackInterface(unsafe.Pointer(header))
	return
}

// MapLen returns the length of the given map interface{} value.
// The provided m must be a map, else it panics.
func MapLen(m any) int {
	if reflect.TypeOf(m).Kind() != reflect.Map {
		panic("reflectx.MapLen: param m must be a map")
	}
	return linkname.Reflect_maplen(EfaceOf(&m).Word)
}

// TypedMemMove exports the typedmemmove function in reflect package.
func TypedMemMove(rtype *RType, dst, src unsafe.Pointer) {
	linkname.Reflect_typedmemmove(unsafe.Pointer(rtype), dst, src)
}

// TypedSliceCopy exports the typedslicecopy function in reflect package.
func TypedSliceCopy(elemRType *RType, dst, src SliceHeader) int {
	return linkname.Reflect_typedslicecopy(unsafe.Pointer(elemRType), dst, src)
}
