package bbp

import (
	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

// OffHeapArena is similar to Arena, except that it allocates memory
// directly from the operating system instead of Go's runtime.
//
// Note that after working with the memory chunks, user **MUST** call
// Free to return the memory to operating system, else memory leaks.
type OffHeapArena Arena

// NewOffHeapArena creates an OffHeapArena which allocates memory directly
// from operating system (without cgo).
// The method Free frees allocated memory chunks.
// Free must be called after working with the arena to avoid memory leaks.
// After Free being called, both the arena and the byte slices allocated
// from the arena **MUST NOT** be touched again.
// chunkSize will be round up to the next power of two that is
// greater than or equal to the system's PAGE_SIZE.
func NewOffHeapArena(chunkSize int) *OffHeapArena {
	chunkSize = alignChunkSize(chunkSize)
	a := arenaPool.Get().(*Arena)
	a.chunkSize = chunkSize
	a.allocFunc = offHeapAlloc
	a.freeFunc = offHeapFree
	return (*OffHeapArena)(a)
}

func offHeapAlloc(chunkSize int) []byte {
	return linkname.Runtime_sysAlloc(uintptr(chunkSize))
}

func offHeapFree(buf []byte) {
	linkname.Runtime_sysFree(buf)
}

// Alloc allocates small byte slice from the arena.
func (a *OffHeapArena) Alloc(length, capacity int) []byte {
	return (*Arena)(a).Alloc(length, capacity)
}

// Free returns all memory chunks managed by the arena to the operating system.
func (a *OffHeapArena) Free() {
	(*Arena)(a).Free()
}
