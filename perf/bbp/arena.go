package bbp

import (
	"container/list"
	"sync"
	"syscall"

	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

var sysPageSize = syscall.Getpagesize()

func alignChunkSize(chunkSize int) int {
	if chunkSize < sysPageSize {
		chunkSize = sysPageSize
	}
	return int(internal.NextPowerOfTwo(uint(chunkSize)))
}

var arenaPool = sync.Pool{
	New: func() any { return &Arena{} },
}

// Arena allocates memory in chunk mode, and serves requests to allocate
// small byte slices, after working with the memory chunks,
// user should call Free to release the allocated memory together.
// It's efficient for memory allocation-heavy workloads.
type Arena struct {
	chunkSize int
	allocFunc func(size int) []byte
	freeFunc  func([]byte)
	lst       list.List
}

// OffHeapArena is similar to Arena, except that it allocates memory
// directly from operating system instead of Go's runtime.
//
// Note that after working with the memory chunks, user **MUST** call
// Free to return the memory to operating system, else memory leaks.
type OffHeapArena Arena

// NewArena creates an Arena object, it allocates memory from the sized
// buffer pools.
// The method Free returns memory chunks to the pool for reusing,
// after which both the arena and the byte slices allocated from the arena
// **MUST NOT** be touched again.
// chunkSize will be round up to the next power of two that is
// greater than or equal to the system's PAGE_SIZE.
func NewArena(chunkSize int) *Arena {
	chunkSize = alignChunkSize(chunkSize)
	poolIdx := indexGet(chunkSize)
	bp := sizedPools[poolIdx]
	a := arenaPool.Get().(*Arena)
	a.chunkSize = chunkSize
	a.allocFunc = bp.Get
	a.freeFunc = bp.Put
	return a
}

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
func (a *Arena) Alloc(length, capacity int) []byte {
	if capacity > a.chunkSize>>2 {
		return make([]byte, length, capacity)
	}

	if active := a.lst.Back(); active != nil {
		chunk := active.Value.(*memChunk)
		if buf, ok := chunk.alloc(length, capacity); ok {
			return buf
		}
	}

	chunk := a.allocNewChunk()
	buf, _ := chunk.alloc(length, capacity)
	return buf
}

// Free releases all memory chunks managed by the arena.
// It returns the memory chunks to pool for reusing.
func (a *Arena) Free() {
	for node := a.lst.Front(); node != nil; node = node.Next() {
		chunk := node.Value.(*memChunk)
		a.freeFunc(chunk.buf)
	}
	a.lst.Init() // clear the list
	arenaPool.Put(a)
}

func (a *Arena) allocNewChunk() *memChunk {
	buf := a.allocFunc(a.chunkSize)
	chunk := &memChunk{buf: buf}
	a.lst.PushBack(chunk)
	return chunk
}

type memChunk struct {
	buf []byte
	i   int
}

func (c *memChunk) alloc(length, capacity int) ([]byte, bool) {
	j := c.i + capacity
	if j < cap(c.buf) {
		c.i = j
		buf := c.buf[j-capacity : j]
		return buf[0:length:capacity], true
	}
	return nil, false
}

// Alloc allocates small byte slice from the arena.
func (a *OffHeapArena) Alloc(length, capacity int) []byte {
	return (*Arena)(a).Alloc(length, capacity)
}

// Free returns all memory chunks managed by the arena to the operating system.
func (a *OffHeapArena) Free() {
	(*Arena)(a).Free()
}
