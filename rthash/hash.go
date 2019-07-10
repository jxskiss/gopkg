// Package rthash exposes the fast hash functions in runtime package.
package rthash

import (
	"reflect"
	"unsafe"
)

// Hash returns a hash code for a comparable argument using a hash function
// that is local to the current invocation of the program.
//
// Hash simply exposes the hash functions in runtime package. It will panic
// if x is nil or unsupported type by the runtime package.
//
// Each call to Local with the same value will return the same result within the
// lifetime of the program, but each run of the program may return different
// results from previous runs.
func Hash(x interface{}) uintptr {
	switch v := x.(type) {
	case string:
		return String(v)
	case []byte:
		return Bytes(v)
	case int8:
		return memhash8(noescape(unsafe.Pointer(&v)), seed)
	case uint8:
		return memhash8(noescape(unsafe.Pointer(&v)), seed)
	case int16:
		return memhash16(noescape(unsafe.Pointer(&v)), seed)
	case uint16:
		return memhash16(noescape(unsafe.Pointer(&v)), seed)
	case int32:
		return Uint32(uint32(v))
	case uint32:
		return Uint32(v)
	case int64:
		return Uint64(uint64(v))
	case uint64:
		return Uint64(v)
	case int:
		return uintptrHash(uintptr(v))
	case uint:
		return uintptrHash(uintptr(v))
	case uintptr:
		return uintptrHash(v)
	case float32:
		return Float32(v)
	case float64:
		return Float64(v)
	case complex64:
		return Complex64(v)
	case complex128:
		return Complex128(v)
	default:
		panic("unsupported hash type")
	}
}

func String(x string) uintptr {
	return stringHash(x, seed)
}

func Bytes(x []byte) uintptr {
	return bytesHash(x, seed)
}

func Int32(x int32) uintptr {
	return int32Hash(uint32(x), seed)
}

func Uint32(x uint32) uintptr {
	return int32Hash(x, seed)
}

func Int64(x int64) uintptr {
	return int64Hash(uint64(x), seed)
}

func Uint64(x uint64) uintptr {
	return int64Hash(x, seed)
}

func Int(x int) uintptr {
	return uintptrHash(uintptr(x))
}

func Uint(x uint) uintptr {
	return uintptrHash(uintptr(x))
}

func Uintptr(x uintptr) uintptr {
	return uintptrHash(x)
}

func Float32(x float32) uintptr {
	return f32hash(noescape(unsafe.Pointer(&x)), seed)
}

func Float64(x float64) uintptr {
	return f64hash(noescape(unsafe.Pointer(&x)), seed)
}

func Complex64(x complex64) uintptr {
	return c64hash(noescape(unsafe.Pointer(&x)), seed)
}

func Complex128(x complex128) uintptr {
	return c128hash(noescape(unsafe.Pointer(&x)), seed)
}

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

//go:linkname getRandomData runtime.getRandomData
func getRandomData(r []byte)

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

// intSize is the size in bits of an int or uint value.
const intSize = 32 << (^uint(0) >> 63)

var seed uintptr

var uintptrHash func(uintptr) uintptr

func init() {
	var tmp []byte
	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&tmp))
	ptr.Data = uintptr(unsafe.Pointer(&seed))
	ptr.Cap = int(unsafe.Sizeof(seed))
	ptr.Len = ptr.Cap
	getRandomData(tmp)

	if intSize == 32 {
		uintptrHash = func(x uintptr) uintptr { return int32Hash(uint32(x), seed) }
	} else {
		uintptrHash = func(x uintptr) uintptr { return int64Hash(uint64(x), seed) }
	}
}
