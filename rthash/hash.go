// Package rthash exposes the various hash functions in runtime package.
//
// The idea mainly comes from https://github.com/golang/go/issues/21195.
package rthash

import (
	"github.com/jxskiss/gopkg/internal/linkname"
	"unsafe"
)

// Hash exposes the various hash functions in runtime package.
// The idea mainly comes from https://github.com/golang/go/issues/21195.
//
// See also: hash/maphash.Hash.
//
// Unlike hash.Hash or hash/maphash.Hash, this Hash does not provide
// the ability to reset seed, the seed must be provided when creating the
// Hash instance and will be used during the lifetime.
//
// This Hash type is intended to be used to do fast sharding, when
// implementing hash tables or other data structures, it's recommended
// to consider using hash/maphash.Hash as a proper choice.
//
// The hash functions are not cryptographically secure.
// (See crypto/sha256 and crypto/sha512 for cryptographic use.)
//
// A Hash must be initialized by calling New().
// After initialized, a Hash is safe for concurrent use by multiple goroutines.
//
// Each call to a same method with the same value will return the same
// result for a Hash instance, but it may and supposed to return different
// hash results from each Hash instance.
type Hash struct {
	seed uintptr
}

// New returns a new Hash instance, which exposes the various hash functions
// in runtime package. The returned Hash instance is safe for concurrent use
// by multiple goroutines.
func New() Hash {
	seed := uintptr(makeseed())
	return Hash{seed: seed}
}

// Hash returns a hash code for a comparable argument.
//
// Note this function calls the hash functions for the concrete type if
// x is of type string, int8, uint8, int16, uint16, int32, uint32,
// int64, uint64, int, uint, uintptr, float32, float64, complex64,
// or complex128, else it calls
func (h Hash) Hash(x interface{}) uintptr {
	switch v := x.(type) {
	case string:
		return h.String(v)
	case int8:
		return h.Int8(v)
	case uint8:
		return h.Uint8(v)
	case int16:
		return h.Int16(v)
	case uint16:
		return h.Uint16(v)
	case int32:
		return h.Int32(v)
	case uint32:
		return h.Uint32(v)
	case int64:
		return h.Int64(v)
	case uint64:
		return h.Uint64(v)
	case int:
		return h.Int(v)
	case uint:
		return h.Uint(v)
	case uintptr:
		return h.Uintptr(v)
	case float32:
		return h.Float32(v)
	case float64:
		return h.Float64(v)
	case complex64:
		return h.Complex64(v)
	case complex128:
		return h.Complex128(v)
	default:
		return h.Interface(v)
	}
}

// String exposes the stringHash function from runtime package.
func (h Hash) String(x string) uintptr {
	return linkname.Runtime_stringHash(x, h.seed)
}

// Bytes exposes the bytesHash function from runtime package.
func (h Hash) Bytes(x []byte) uintptr {
	return linkname.Runtime_bytesHash(x, h.seed)
}

// Int8 exposes the memhash8 function from runtime package.
func (h Hash) Int8(x int8) uintptr {
	return linkname.Runtime_memhash8(unsafe.Pointer(&x), h.seed)
}

// Uint8 exposes the memhash8 function from runtime package.
func (h Hash) Uint8(x uint8) uintptr {
	return linkname.Runtime_memhash8(unsafe.Pointer(&x), h.seed)
}

// Int16 exposes the memhash16 function from runtime package.
func (h Hash) Int16(x int16) uintptr {
	return linkname.Runtime_memhash16(unsafe.Pointer(&x), h.seed)
}

// Uint16 exposes the memhash16 function from runtime package.
func (h Hash) Uint16(x uint16) uintptr {
	return linkname.Runtime_memhash16(unsafe.Pointer(&x), h.seed)
}

// Int32 exposes the int32Hash function from runtime package.
func (h Hash) Int32(x int32) uintptr {
	return linkname.Runtime_int32Hash(uint32(x), h.seed)
}

// Uint32 exposes the int32Hash function from runtime package.
func (h Hash) Uint32(x uint32) uintptr {
	return linkname.Runtime_int32Hash(x, h.seed)
}

// Int64 exposes the int64Hash function from runtime package.
func (h Hash) Int64(x int64) uintptr {
	return linkname.Runtime_int64Hash(uint64(x), h.seed)
}

// Uint64 exposes the int64Hash function from runtime package.
func (h Hash) Uint64(x uint64) uintptr {
	return linkname.Runtime_int64Hash(x, h.seed)
}

// Int calculates hash of x using either int32Hash or int64Hash
// according to the pointer size of the platform.
func (h Hash) Int(x int) uintptr {
	if ptrSize == 32 {
		return linkname.Runtime_int32Hash(uint32(x), h.seed)
	}
	return linkname.Runtime_int64Hash(uint64(x), h.seed)
}

// Uint calculates hash of x using either int32Hash or int64Hash
// according the pointer size of the platform.
func (h Hash) Uint(x uint) uintptr {
	if ptrSize == 32 {
		return linkname.Runtime_int32Hash(uint32(x), h.seed)
	}
	return linkname.Runtime_int64Hash(uint64(x), h.seed)
}

// Uintptr calculates hash of x using either int32Hash or int64Hash
// according to the pointer size of the platform.
func (h Hash) Uintptr(x uintptr) uintptr {
	if ptrSize == 32 {
		return linkname.Runtime_int32Hash(uint32(x), h.seed)
	}
	return linkname.Runtime_int64Hash(uint64(x), h.seed)
}

// Float32 exposes the f32hash function from runtime package.
func (h Hash) Float32(x float32) uintptr {
	return linkname.Runtime_f32hash(unsafe.Pointer(&x), h.seed)
}

// Float64 exposes the f64hash function from runtime package.
func (h Hash) Float64(x float64) uintptr {
	return linkname.Runtime_f64hash(unsafe.Pointer(&x), h.seed)
}

// Complex64 exposes the c64hash function from runtime package.
func (h Hash) Complex64(x complex64) uintptr {
	return linkname.Runtime_c64hash(unsafe.Pointer(&x), h.seed)
}

// Complex128 exposes the c128hash function from runtime package.
func (h Hash) Complex128(x complex128) uintptr {
	return linkname.Runtime_c128hash(unsafe.Pointer(&x), h.seed)
}

// Interface exposes the efaceHash function from runtime package.
func (h Hash) Interface(x interface{}) uintptr {
	return linkname.Runtime_efaceHash(x, h.seed)
}

// ptrSize is the size in bits of an int or uint value.
const ptrSize = 32 << (^uint(0) >> 63)

func makeseed() uint64 {
	var s1, s2 uint64
	for {
		s1 = uint64(linkname.Runtime_fastrand())
		s2 = uint64(linkname.Runtime_fastrand())
		// We use seed 0 to indicate an uninitialized seed/hash,
		// so keep trying until we get a non-zero seed.
		if s1|s2 != 0 {
			break
		}
	}
	return s1<<32 + s2
}
