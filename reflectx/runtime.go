package reflectx

import (
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
type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// SliceHeader is the runtime representation of a slice.
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

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
// f with each pair of key value pointers.
func MapIter(m interface{}, f func(k, v unsafe.Pointer)) {
	eface := EFaceOf(&m)
	iter := mapiterinit(eface.RType, eface.Word)
	for iter.key != nil {
		f(iter.key, iter.value)
		mapiternext(iter)
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
	header = &SliceHeader{
		Data: unsafe_NewArray(ToRType(elemTyp), capacity),
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
	typedmemmove(rtype, dst, src)
}

// TypedSliceCopy exports the typedslicecopy function in reflect package.
func TypedSliceCopy(elemRType *RType, dst, src SliceHeader) int {
	return typedslicecopy(elemRType, dst, src)
}

// ------------------------------------------------------------ //

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(*RType) unsafe.Pointer

//go:linkname unsafe_NewArray reflect.unsafe_NewArray
func unsafe_NewArray(*RType, int) unsafe.Pointer

// typedmemmove copies a value of type t to dst from src.
//go:noescape
//go:linkname typedmemmove reflect.typedmemmove
func typedmemmove(t *RType, dst, src unsafe.Pointer)

// typedslicecopy copies a slice of elemType values from src to dst,
// returning the number of elements copied.
//go:noescape
//go:linkname typedslicecopy reflect.typedslicecopy
func typedslicecopy(elemRType *RType, dst, src SliceHeader) int

//go:noescape
//go:linkname maplen reflect.maplen
func maplen(m unsafe.Pointer) int

// m escapes into the return value, but the caller of mapiterinit
// doesn't let the return value escape.
//go:noescape
//go:linkname mapiterinit reflect.mapiterinit
func mapiterinit(rtype *RType, m unsafe.Pointer) *hiter

//go:noescape
//go:linkname mapiternext reflect.mapiternext
func mapiternext(it *hiter)

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
