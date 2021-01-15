package reflectx

import (
	"unsafe"
)

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(unsafe.Pointer) unsafe.Pointer

//go:linkname unsafe_NewArray reflect.unsafe_NewArray
func unsafe_NewArray(unsafe.Pointer, int) unsafe.Pointer

// typedmemmove copies a value of type t to dst from src.
//go:noescape
//go:linkname typedmemmove reflect.typedmemmove
func typedmemmove(t unsafe.Pointer, dst, src unsafe.Pointer)

//go:noescape
//go:linkname maplen reflect.maplen
func maplen(m unsafe.Pointer) int

// m escapes into the return value, but the caller of mapiterinit
// doesn't let the return value escape.
//go:noescape
//go:linkname mapiterinit reflect.mapiterinit
func mapiterinit(rtype unsafe.Pointer, m unsafe.Pointer) *hiter

//go:noescape
//go:linkname mapiternext reflect.mapiternext
func mapiternext(it *hiter)

// A hash iteration structure.
// If you modify hiter, also change cmd/internal/gc/reflect.go to indicate
// the layout of this structure.
type hiter struct {
	key   unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/internal/gc/range.go).
	value unsafe.Pointer // Must be in second position (see cmd/internal/gc/range.go).
	// rest fields are ignored
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	RType unsafe.Pointer
	Word  unsafe.Pointer
}

// value is the reflection data to a Go value.
// See reflect/value.go#Value for more details.
type value struct {
	typ  unsafe.Pointer
	ptr  unsafe.Pointer
	flag uintptr
}

const (
	// IntBitSize is the size in bits of an int or uint value.
	IntBitSize = 32 << (^uint(0) >> 63)

	// IntByteSize is the size in bytes of an int or uint values.
	IntByteSize = IntBitSize / 8

	IsPlatform32bit = IntBitSize == 32
	IsPlatform64bit = IntBitSize == 64
)
