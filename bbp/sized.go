package bbp

import (
	"math/bits"
	"sync"
)

const (
	// at least 1<<6 = 64B, a CPU cache line size
	minPoolIdx = 6

	// max 1<<25 = 32MB
	poolSize = 26
)

const (
	// minSize is the minimum buffer size (64B) provided in this package.
	minSize = 1 << minPoolIdx

	// maxSize is the maximum buffer size (32MB) provided in this package.
	maxSize = 1 << (poolSize - 1)
)

// Get returns a byte buffer from the pool with specified length and capacity.
// The returned byte buffer's capacity is at least 64.
//
// The returned byte buffer can be put back to the pool by calling Put(buf),
// which may be reused later. This reduces memory allocations and GC pressure.
func Get(length int, capacity int) *Buffer {
	if capacity > maxSize {
		return &Buffer{
			buf: make([]byte, length, capacity),
		}
	}
	buf := &Buffer{
		buf: get(length, capacity),
	}
	return buf
}

// Put puts back a byte buffer to the pool for reusing.
//
// The buf mustn't be touched after retuning it to the pool.
// Otherwise, data races will occur.
func Put(buf *Buffer) {
	put(buf.buf)
}

// Grow returns a new byte buffer from the pool which guarantees it's
// at least of specified capacity.
//
// If cap(buf) >= capacity, it returns buf directly, else it returns a
// new byte buffer with data of buf copied.
func Grow(buf []byte, capacity int) []byte {
	if cap(buf) >= capacity {
		return buf
	}
	return grow(buf, capacity, false)
}

// PutSlice puts back a byte slice to the pool.
//
// The byte slice mustn't be touched after returning it to the pool,
// otherwise data races will occur.
func PutSlice(buf []byte) { put(buf) }

// -------- sized pools -------- //

// power of two sized pools
var sizedPools [poolSize]sync.Pool

func init() {
	for i := 0; i < poolSize; i++ {
		size := 1 << i
		sizedPools[i].New = func() interface{} {
			buf := make([]byte, 0, size)
			return buf
		}
	}
}

// callers must guarantee that capacity is not greater than maxSize.
func get(length, capacity int) []byte {
	idx := indexGet(capacity)
	out := sizedPools[idx].Get().([]byte)
	return out[:length]
}

func put(buf []byte) {
	cap_ := cap(buf)
	if cap_ >= minSize && cap_ <= maxSize {
		idx := indexPut(cap_)
		buf = buf[:0]
		sizedPools[idx].Put(buf)
	}
}

func grow(buf []byte, capacity int, reuse bool) []byte {
	var newBuf []byte
	if capacity > maxSize {
		newBuf = make([]byte, len(buf), capacity)
	} else {
		newBuf = get(len(buf), capacity)
	}
	copy(newBuf, buf)
	if reuse {
		put(buf)
	}
	return newBuf
}

// indexGet finds the pool index for the given size to get buffer from,
// if size is not power of tow, it returns the next power of tow index.
func indexGet(size int) int {
	if size <= minSize {
		return minPoolIdx
	}
	idx := bsr(size)
	if isPowerOfTwo(size) {
		return idx
	}
	return idx + 1
}

// indexPut finds the pool index for the given size to put buffer back,
// if size is not power of two, it returns the previous power of two index.
func indexPut(size int) int {
	return bsr(size)
}

// bsr.
//
// Callers within this package guarantee that n doesn't overflow int32.
func bsr(n int) int {
	return bits.Len32(uint32(n)) - 1
}

// isPowerOfTwo reports whether n is a power of two.
func isPowerOfTwo(n int) bool {
	return n&(n-1) == 0
}
