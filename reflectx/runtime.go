package reflectx

import (
	"github.com/jxskiss/gopkg/internal/linkname"
	"reflect"
	"unsafe"
)

func add(p unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + offset)
}

// ArrayAt returns the i-th element of p,
// an array whose elements are elemSize bytes wide.
// The array pointed at by p must have at least i+1 elements:
// it is invalid (but impossible to check here) to pass i >= len,
// because then the result will point outside the array.
func ArrayAt(p unsafe.Pointer, i int, elemSize uintptr) unsafe.Pointer {
	return add(p, uintptr(i)*elemSize)
}

// MakeSlice makes a new slice of the given reflect.Type and length, capacity.
func MakeSlice(elemTyp reflect.Type, length, capacity int) (slice interface{}, header *SliceHeader) {
	elemRType := ToRType(elemTyp)
	data := linkname.Reflect_unsafe_NewArray(unsafe.Pointer(elemRType), capacity)
	header = &SliceHeader{
		Data: data,
		Len:  length,
		Cap:  capacity,
	}
	slice = *(*interface{})(unsafe.Pointer(&EmptyInterface{
		RType: SliceOf(elemRType),
		Word:  unsafe.Pointer(header),
	}))
	return
}

// MapLen returns the length of the given map interface{} value.
// The provided m must be a map, else it panics.
func MapLen(m interface{}) int {
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
