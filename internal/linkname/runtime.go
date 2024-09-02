package linkname

import "unsafe"

// Pid returns the id of current p.
//
//go:nosplit
func Pid() int {
	pid := runtime_procPin()
	runtime_procUnpin()
	return pid
}

//go:linkname runtime_procPin runtime.procPin
//go:nosplit
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
//go:nosplit
func runtime_procUnpin()

// -------- runtime hash functions --------

//go:linkname Runtime_memhash32 runtime.memhash32
func Runtime_memhash32(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_memhash64 runtime.memhash64
func Runtime_memhash64(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_stringHash runtime.stringHash
func Runtime_stringHash(s string, seed uintptr) uintptr

//go:linkname Runtime_nilinterhash runtime.nilinterhash
func Runtime_nilinterhash(p unsafe.Pointer, h uintptr) uintptr

//go:linkname Runtime_typehash runtime.typehash
func Runtime_typehash(rtype unsafe.Pointer, p unsafe.Pointer, h uintptr) uintptr

// -------- runtime malloc without memclr cost --------

//go:linkname Runtime_mallocgc runtime.mallocgc
func Runtime_mallocgc(size uintptr, typ unsafe.Pointer, needzero bool) unsafe.Pointer
