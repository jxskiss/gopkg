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

//go:linkname Runtime_bytesHash runtime.bytesHash
func Runtime_bytesHash(b []byte, seed uintptr) uintptr

//go:linkname Runtime_efaceHash runtime.efaceHash
func Runtime_efaceHash(i any, seed uintptr) uintptr

//go:linkname Runtime_typehash runtime.typehash
func Runtime_typehash(rtype unsafe.Pointer, p unsafe.Pointer, h uintptr) uintptr

// -------- runtime moduledata --------

//go:linkname Runtime_activeModules runtime.activeModules
func Runtime_activeModules() []unsafe.Pointer
