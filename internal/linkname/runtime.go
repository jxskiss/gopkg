//go:build gc && !go1.22

package linkname

import (
	_ "runtime"
	"unsafe"
)

var sysAllocMemStat uint64

// Runtime_memclrNoHeapPointers clears n bytes starting at ptr.
//
// Usually you should use typedmemclr. Runtime_memclrNoHeapPointers should be
// used only when the caller knows that *ptr contains no heap pointers
// because either:
//
// *ptr is initialized memory and its type is pointer-free, or
//
// *ptr is uninitialized memory (e.g., memory that's being reused
// for a new allocation) and hence contains only "junk".
//
// Runtime_memclrNoHeapPointers ensures that if ptr is pointer-aligned, and n
// is a multiple of the pointer size, then any pointer-aligned,
// pointer-sized portion is cleared atomically. Despite the function
// name, this is necessary because this function is the underlying
// implementation of typedmemclr and memclrHasPointers. See the doc of
// memmove for more details.
//
// The (CPU-specific) implementations of this function are in memclr_*.s.
//
//go:noescape
//go:linkname Runtime_memclrNoHeapPointers runtime.memclrNoHeapPointers
func Runtime_memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)

//go:noescape
//go:linkname Runtime_fastrand runtime.fastrand
func Runtime_fastrand() uint32

//go:noescape
//go:linkname Runtime_fastrandn runtime.fastrandn
func Runtime_fastrandn(n uint32) uint32

// Runtime_procPin links to runtime.procPin.
// It pins current p, return pid.
//
// DON'T USE THIS if you don't know what it is.
//
//go:noescape
//go:linkname Runtime_procPin runtime.procPin
func Runtime_procPin() int

// Runtime_procUnpin links to runtime.procUnpin.
// It unpins current p.
//
//go:noescape
//go:linkname Runtime_procUnpin runtime.procUnpin
func Runtime_procUnpin()

// Pid returns the id of current p.
func Pid() int {
	pid := Runtime_procPin()
	Runtime_procUnpin()
	return pid
}

//go:linkname Runtime_memhash8 runtime.memhash8
func Runtime_memhash8(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_memhash16 runtime.memhash16
func Runtime_memhash16(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_stringHash runtime.stringHash
func Runtime_stringHash(s string, seed uintptr) uintptr

//go:linkname Runtime_bytesHash runtime.bytesHash
func Runtime_bytesHash(b []byte, seed uintptr) uintptr

//go:linkname Runtime_int32Hash runtime.int32Hash
func Runtime_int32Hash(i uint32, seed uintptr) uintptr

//go:linkname Runtime_int64Hash runtime.int64Hash
func Runtime_int64Hash(i uint64, seed uintptr) uintptr

//go:linkname Runtime_f32hash runtime.f32hash
func Runtime_f32hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_f64hash runtime.f64hash
func Runtime_f64hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_c64hash runtime.c64hash
func Runtime_c64hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_c128hash runtime.c128hash
func Runtime_c128hash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_efaceHash runtime.efaceHash
func Runtime_efaceHash(i any, seed uintptr) uintptr

//go:linkname Runtime_typehash runtime.typehash
func Runtime_typehash(rtype unsafe.Pointer, p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_activeModules runtime.activeModules
func Runtime_activeModules() []unsafe.Pointer

//go:linkname sysBigEndian internal.sys.BigEndian
var sysBigEndian bool

// Runtime_readUnaligned32 reads memory pointed by p as a uint32 value.
// It performs the read with a native endianness.
//
// It is copied from runtime.readUnaligned32 instead of linked to help
// inlining.
func Runtime_readUnaligned32(p unsafe.Pointer) uint32 {
	q := (*[4]byte)(p)
	if sysBigEndian {
		return uint32(q[3]) | uint32(q[2])<<8 | uint32(q[1])<<16 | uint32(q[0])<<24
	}
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16 | uint32(q[3])<<24
}

// Runtime_readUnaligned64 reads memory pointed by p as a uint64 value.
// It performs the read with a native endianness.
//
// It is copied from runtime.readUnaligned64 but instead of linked to
// help inlining.
func Runtime_readUnaligned64(p unsafe.Pointer) uint64 {
	q := (*[8]byte)(p)
	if sysBigEndian {
		return uint64(q[7]) | uint64(q[6])<<8 | uint64(q[5])<<16 | uint64(q[4])<<24 |
			uint64(q[3])<<32 | uint64(q[2])<<40 | uint64(q[1])<<48 | uint64(q[0])<<56
	}
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24 |
		uint64(q[4])<<32 | uint64(q[5])<<40 | uint64(q[6])<<48 | uint64(q[7])<<56
}
