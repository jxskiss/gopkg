package linkname

import (
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

func Reflect_typelinks() ([]unsafe.Pointer, [][]int32) {
	return reflect_typelinks()
}

func Reflect_resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer {
	return reflect_resolveTypeOff(rtype, off)
}

func Reflect_ifaceIndir(rtype unsafe.Pointer) bool {
	return reflect_ifaceIndir(rtype)
}

func Reflect_unsafe_New(rtype unsafe.Pointer) unsafe.Pointer {
	return reflect_unsafe_New(rtype)
}

func Reflect_unsafe_NewArray(elemRType unsafe.Pointer, capacity int) unsafe.Pointer {
	return reflect_unsafe_NewArray(elemRType, capacity)
}

// Reflect_typedmemmove copies a value of type t to dst from src.
func Reflect_typedmemmove(rtype unsafe.Pointer, dst, src unsafe.Pointer) {
	reflect_typedmemmove(rtype, dst, src)
}

// Reflect_typedslicecopy copies a slice of elemType values from src to dst,
// returning the number of elements copied.
func Reflect_typedslicecopy(elemRType unsafe.Pointer, dst, src unsafeheader.SliceHeader) int {
	return reflect_typedslicecopy(elemRType, dst, src)
}

func Reflect_maplen(m unsafe.Pointer) int {
	return reflect_maplen(m)
}

//go:linkname reflect_typelinks reflect.typelinks
func reflect_typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname reflect_resolveTypeOff reflect.resolveTypeOff
func reflect_resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer

//go:linkname reflect_ifaceIndir reflect.ifaceIndir
//go:noescape
func reflect_ifaceIndir(rtype unsafe.Pointer) bool

//go:linkname reflect_unsafe_New reflect.unsafe_New
func reflect_unsafe_New(unsafe.Pointer) unsafe.Pointer

//go:linkname reflect_unsafe_NewArray reflect.unsafe_NewArray
func reflect_unsafe_NewArray(unsafe.Pointer, int) unsafe.Pointer

//go:linkname reflect_typedmemmove reflect.typedmemmove
//go:noescape
func reflect_typedmemmove(t unsafe.Pointer, dst, src unsafe.Pointer)

//go:linkname reflect_typedslicecopy reflect.typedslicecopy
//go:noescape
func reflect_typedslicecopy(elemRType unsafe.Pointer, dst, src unsafeheader.SliceHeader) int

//go:linkname reflect_maplen reflect.maplen
//go:noescape
func reflect_maplen(m unsafe.Pointer) int
