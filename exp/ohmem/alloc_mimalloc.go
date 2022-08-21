//go:build mimalloc

package ohmem

/*
#cgo CFLAGS: -I${SRCDIR}/mimalloc-1.7.6/include -I${SRCDIR}/mimalloc-1.7.6 -O3

#include <mimalloc.h>
#include <src/static.c>
*/
import "C"
import (
	"reflect"
	"unsafe"
)

func _C_zalloc(n int) []byte {
	ptr := C.mi_zalloc(C.size_t(n))
	if ptr == nil {
		// NB: throw is like panic, except it guarantees the process will be
		// terminated. The call below is exactly what the Go runtime invokes when
		// it cannot allocate memory.
		throw("out of memory")
	}
	return _getBytes(uintptr(ptr), n, n)
}

func _C_free(mem []byte) {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&mem))
	C.mi_free(unsafe.Pointer(h.Data))
}
