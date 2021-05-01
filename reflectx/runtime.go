package reflectx

import (
	"github.com/jxskiss/gopkg/internal/linkname"
	"github.com/jxskiss/gopkg/internal/unsafeheader"
	"reflect"
	"unsafe"
)

const (
	// PtrBitSize is the size in bits of an int or uint value.
	PtrBitSize = 32 << (^uint(0) >> 63)

	// PtrByteSize is the size in bytes of an int or uint values.
	PtrByteSize = PtrBitSize / 8

	IsPlatform32bit = PtrBitSize == 32
	IsPlatform64bit = PtrBitSize == 64
)

// StringHeader is the runtime representation of a string.
// Unlike reflect.StringHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type StringHeader = unsafeheader.String

// SliceHeader is the runtime representation of a slice.
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type SliceHeader = unsafeheader.Slice

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	sh := (*StringHeader)(unsafe.Pointer(&s))
	bh := &SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}

// EmptyInterface is the header for an interface{} value.
// It's a copy type of runtime.eface.
type EmptyInterface struct {
	RType *RType // *rtype
	Word  unsafe.Pointer
}

// EFaceOf casts the empty interface{} pointer to an EmptyInterface pointer.
func EFaceOf(ep *interface{}) *EmptyInterface {
	return (*EmptyInterface)(unsafe.Pointer(ep))
}

// PackInterface pack an empty interface{} using the given reflect.Type
// and data pointer.
func PackInterface(typ reflect.Type, word unsafe.Pointer) interface{} {
	return ToRType(typ).PackInterface(word)
}

// MapLen returns the length of the given map interface{} value.
// The provided m must be a map, else it panics.
func MapLen(m interface{}) int {
	return maplen(EFaceOf(&m).Word)
}

// MapIter iterates the given map interface{} value, and calls function
// f with each pair of key value interface{}.
// The iteration can be aborted by returning a non-zero value from f.
func MapIter(m interface{}, f func(k, v interface{}) int) {
	eface := EFaceOf(&m)
	keyTyp := eface.RType.Key()
	elemTyp := eface.RType.Elem()
	iter := mapiterinit(eface.RType, eface.Word)
	for iter.key != nil {
		k := keyTyp.PackInterface(iter.key)
		v := elemTyp.PackInterface(iter.value)
		if f(k, v) != 0 {
			return
		}
		mapiternext(iter)
	}
}

// MapIterPointer is similar to MapIter, but it calls f for each key
// value pair with their address pointers.
// The iteration can be aborted by returning a non-zero value from f.
func MapIterPointer(m interface{}, f func(k, v unsafe.Pointer) int) {
	eface := EFaceOf(&m)
	iter := mapiterinit(eface.RType, eface.Word)
	for iter.key != nil {
		n := f(iter.key, iter.value)
		if n != 0 {
			return
		}
		mapiternext(iter)
	}
}

// SliceLen returns the length of the given slice interface{} value.
// The provided slice must be a slice, else it panics.
func SliceLen(slice interface{}) int {
	return UnpackSlice(slice).Len
}

// SliceCap returns the capacity of the given slice interface{} value.
// The provided slice must be a slice, else it panics.
func SliceCap(slice interface{}) int {
	return UnpackSlice(slice).Cap
}

// SliceIter iterates the given slice interface{} value, and calls
// function f with each element in the slice.
// The iteration can be aborted by returning a non-zero value from f.
func SliceIter(slice interface{}, f func(elem interface{}) int) {
	elemTyp := EFaceOf(&slice).RType.Elem()
	elemSize := elemTyp.Size()
	header := UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		ptr := ArrayAt(header.Data, i, elemSize)
		elem := elemTyp.PackInterface(ptr)
		if f(elem) != 0 {
			return
		}
	}
}

// SliceIterPointer is similar to SliceIter, but it calls f for each
// element with the address pointer.
// The iteration can be aborted by returning a non-zero value from f.
func SliceIterPointer(slice interface{}, f func(elem unsafe.Pointer) int) {
	elemTyp := EFaceOf(&slice).RType.Elem()
	elemSize := elemTyp.Size()
	header := UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		elem := ArrayAt(header.Data, i, elemSize)
		n := f(elem)
		if n != 0 {
			return
		}
	}
}

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

// UnpackSlice unpacks the given slice interface{} to a SliceHeader.
func UnpackSlice(slice interface{}) SliceHeader {
	return *(*SliceHeader)(EFaceOf(&slice).Word)
}

// CastSlice returns a new interface{} value of the given reflect.Type,
// and the data pointer points to the underlying data of the given slice.
func CastSlice(slice interface{}, typ reflect.Type) interface{} {
	return ToRType(typ).PackInterface(EFaceOf(&slice).Word)
}

// MakeSlice makes a new slice of the given reflect.Type and length, capacity.
func MakeSlice(elemTyp reflect.Type, length, capacity int) (
	slice interface{}, header *SliceHeader,
) {
	elemRType := ToRType(elemTyp)
	data := linkname.Reflect_unsafe_NewArray(unsafe.Pointer(elemRType), capacity)
	header = &SliceHeader{
		Data: data,
		Len:  length,
		Cap:  capacity,
	}
	slice = *(*interface{})(unsafe.Pointer(&EmptyInterface{
		RType: ToRType(reflect.SliceOf(elemTyp)),
		Word:  unsafe.Pointer(header),
	}))
	return
}

// TypedMemMove exposes the typedmemmove function in reflect package.
func TypedMemMove(rtype *RType, dst, src unsafe.Pointer) {
	linkname.Reflect_typedmemmove(unsafe.Pointer(rtype), dst, src)
}

// TypedSliceCopy exports the typedslicecopy function in reflect package.
func TypedSliceCopy(elemRType *RType, dst, src SliceHeader) int {
	return linkname.Reflect_typedslicecopy(unsafe.Pointer(elemRType), dst, src)
}

// ------------------------------------------------------------ //

func maplen(m unsafe.Pointer) int {
	return linkname.Reflect_maplen(m)
}

func mapiterinit(rtype *RType, m unsafe.Pointer) *hiter {
	return (*hiter)(linkname.Reflect_mapiterinit(unsafe.Pointer(rtype), m))
}

func mapiternext(it *hiter) {
	linkname.Reflect_mapiternext(unsafe.Pointer(it))
}

// A hash iteration structure.
// If you modify hiter, also change cmd/internal/gc/reflect.go to indicate
// the layout of this structure.
type hiter struct {
	key   unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/internal/gc/range.go).
	value unsafe.Pointer // Must be in second position (see cmd/internal/gc/range.go).

	// The rest fields are not used within this package.
	// ...
}

// iface is a copy type of runtime.iface.
type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}
