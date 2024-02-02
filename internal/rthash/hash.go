package rthash

import (
	"reflect"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

type HashFunc[K comparable] func(key K) uintptr

// NewHashFunc returns a new hash function, which exposes the various
// hash functions in runtime package.
//
// Note that this function generates a random seed, each calling of this
// function returns DIFFERENT hash function, different hash functions
// generate different hash result for same input.
//
// The returned function is safe for concurrent use by multiple goroutines.
func NewHashFunc[K comparable]() HashFunc[K] {
	var seed uintptr
	for seed == 0 {
		seed = uintptr(linkname.Runtime_fastrand64())
	}

	var zero K
	typ := reflect.TypeOf(zero)
	if typ == nil {
		return func(key K) uintptr {
			return linkname.Runtime_efaceHash(key, seed)
		}
	}

	switch typ.Kind() {
	case reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		size := unsafe.Sizeof(zero)
		if size == 4 {
			return func(key K) uintptr {
				return linkname.Runtime_memhash32(noescape(unsafe.Pointer(&key)), seed)
			}
		}
		return func(key K) uintptr {
			return linkname.Runtime_memhash64(noescape(unsafe.Pointer(&key)), seed)
		}
	case reflect.String:
		return func(key K) uintptr {
			return linkname.Runtime_stringHash(*(*string)(unsafe.Pointer(&key)), seed)
		}
	default:
		rtype := unsafeheader.ToRType(typ)
		return func(key K) uintptr {
			return linkname.Runtime_typehash(rtype, unsafe.Pointer(&key), seed)
		}
	}
}

// NewBytesHash returns a new hash function to hash byte slice.
//
// Note that this function generates a random seed, each calling of this
// function returns DIFFERENT hash function, different hash functions
// generate different hash result for same input.
//
// The returned function is safe for concurrent use by multiple goroutines.
func NewBytesHash() func(b []byte) uintptr {
	var seed uintptr
	for seed == 0 {
		seed = uintptr(linkname.Runtime_fastrand64())
	}
	return func(b []byte) uintptr {
		return linkname.Runtime_bytesHash(b, seed)
	}
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
//
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0) //nolint:staticcheck
}
