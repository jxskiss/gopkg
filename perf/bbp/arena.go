package bbp

import (
	"sync"
	"syscall"

	"github.com/jxskiss/gopkg/v2/internal"
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
	chunks    []memChunk
}

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

// Alloc allocates small byte slice from the arena.
func (a *Arena) Alloc(length, capacity int) []byte {
	if capacity > a.chunkSize>>2 {
		return make([]byte, length, capacity)
	}

	if n := len(a.chunks); n > 0 {
		c := &a.chunks[len(a.chunks)-1]
		if buf, ok := c.alloc(length, capacity); ok {
			return buf
		}
	}

	// Need to alloc new memory chunk.
	newMem := a.allocFunc(a.chunkSize)
	a.chunks = append(a.chunks, memChunk{buf: newMem})
	c := &a.chunks[len(a.chunks)-1]
	buf, _ := c.alloc(length, capacity)
	return buf
}

// Free releases all memory chunks managed by the arena.
// It returns the memory chunks to pool for reusing.
func (a *Arena) Free() {
	for i := range a.chunks {
		if a.chunks[i].buf != nil {
			a.freeFunc(a.chunks[i].buf)
			a.chunks[i].buf = nil
		}
	}
	a.chunks = a.chunks[:0]
	arenaPool.Put(a)
}

type memChunk struct {
	buf []byte
	i   int
}

func (c *memChunk) alloc(length, capacity int) ([]byte, bool) {
	if j := c.i + capacity; j <= cap(c.buf) {
		buf := c.buf[c.i:j]
		c.i = j
		return buf[0:length:capacity], true
	}
	return nil, false
}
