//go:build cgo

package bbp

/*
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// NewCgoArena creates a OffHeapArena which allocates memory by calling
// cgo `C.malloc`. cgo must be enabled to use this.
// The method Free releases allocated memory chunks. Free must be called
// after working with the arena to avoid memory leaks.
// After Free being called, both the arena and the byte slice allocated
// from the arena **MUST NOT** be touched again.
// chunkSize must be greater than or equal to PAGE_SIZE and will be round up
// to the next power of two.
func NewCgoArena(chunkSize int) *OffHeapArena {
	chunkSize = alignChunkSize(chunkSize)
	a := arenaPool.Get().(*Arena)
	a.chunkSize = chunkSize
	a.allocFunc = cgoAlloc
	a.freeFunc = cgoFree
	return (*OffHeapArena)(a)
}

func cgoAlloc(size int) []byte {
	ptr := C.malloc(C.size_t(size))
	if ptr == nil {
		// Don't allow the caller to capture this panic,
		// and block to wait the program exiting.
		go func() {
			panic("bbp.Arena: out of memory")
		}()
		select {}
	}
	buf := *(*[]byte)(unsafe.Pointer(&unsafeheader.Slice{
		Data: ptr,
		Len:  size,
		Cap:  size,
	}))
	return buf
}

func cgoFree(buf []byte) {
	ptr := (*unsafeheader.Slice)(unsafe.Pointer(&buf)).Data
	C.free(ptr)
}
