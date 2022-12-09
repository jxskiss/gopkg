//go:build gc && go1.19 && !go1.21

package linkname

import (
	"reflect"
	"unsafe"
)

//go:noescape
//go:linkname Runtime_fastrand64 runtime.fastrand64
func Runtime_fastrand64() uint64

// Runtime_sysAlloc allocates memory off heap by calling runtime.sysAllocOS.
//
// DON'T use this if you don't know what it does.
func Runtime_sysAlloc(n uintptr) []byte {
	addr := runtime_sysAllocOS(n)
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(addr),
		Len:  int(n),
		Cap:  int(n),
	}))
}

// Runtime_sysFree frees memory allocated by Runtime_sysAlloc.
//
// DON'T use this if you don't know what it does.
func Runtime_sysFree(mem []byte) {
	addr := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&mem)).Data)
	n := uintptr(cap(mem))
	runtime_sysFreeOS(addr, n)
}

//go:linkname runtime_sysAllocOS runtime.sysAllocOS
func runtime_sysAllocOS(n uintptr) unsafe.Pointer

//go:linkname runtime_sysFreeOS runtime.sysFreeOS
func runtime_sysFreeOS(v unsafe.Pointer, n uintptr)
