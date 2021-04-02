package rthash

import "unsafe"

// ptrSize is the size in bits of an int or uint value.
const ptrSize = 32 << (^uint(0) >> 63)

//go:linkname memhash8 runtime.memhash8
func memhash8(p unsafe.Pointer, h uintptr) uintptr

//go:linkname memhash16 runtime.memhash16
func memhash16(p unsafe.Pointer, h uintptr) uintptr

//go:linkname stringHash runtime.stringHash
func stringHash(s string, seed uintptr) uintptr

//go:linkname bytesHash runtime.bytesHash
func bytesHash(b []byte, seed uintptr) uintptr

//go:linkname int32Hash runtime.int32Hash
func int32Hash(i uint32, seed uintptr) uintptr

//go:linkname int64Hash runtime.int64Hash
func int64Hash(i uint64, seed uintptr) uintptr

//go:linkname f32hash runtime.f32hash
func f32hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname f64hash runtime.f64hash
func f64hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname c64hash runtime.c64hash
func c64hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname c128hash runtime.c128hash
func c128hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname efaceHash runtime.efaceHash
func efaceHash(i interface{}, seed uintptr) uintptr

//go:noescape
//go:linkname _fastrand runtime.fastrand
func _fastrand() uint32

// noescape is copied from the runtime package.
//
// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
