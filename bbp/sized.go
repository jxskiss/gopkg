package bbp

import (
	"math/bits"
	"sync"
)

const (
	// MinBufSize is the minimum buffer size (8B) provided in this package.
	MinBufSize = 1 << minPoolIdx

	// MaxBufSize is the maximum buffer size (64MB) provided in this package.
	MaxBufSize = 1 << (poolSize - 1)
)

// Get returns a byte buffer from the pool with specified length and capacity.
// The returned byte buffer's capacity is of at least 8.
//
// The returned byte buffer can be put back to the pool by calling Put(buf),
// which may be reused later. This reduces memory allocations and GC pressure.
func Get(length int, capacity ...int) *Buffer {
	cap_ := length
	if len(capacity) == 1 {
		cap_ = capacity[0]
	} else if len(capacity) > 1 {
		panic("too many arguments to bbp.Get")
	}
	if cap_ > MaxBufSize {
		return &Buffer{
			B:       make([]byte, length, cap_),
			noReuse: true,
		}
	}
	buf := getBuffer()
	buf.B = get(length, cap_)
	return buf
}

// Grow returns a new byte buffer from the pool which guarantees it's
// at least of specified capacity.
//
// If capacity is not specified, the returned slice is at least twice
// of the given buf slice.
// The returned byte buffer's capacity is always power of two, which
// can be put back to the pool after usage.
//
// The old buf will be put into the pool for reusing, so it mustn't be
// touched after calling this function, otherwise data races will occur.
func Grow(buf []byte, capacity ...int) []byte {
	if len(capacity) > 1 {
		panic("too many arguments to bbp.Grow")
	}
	l, c := len(buf), cap(buf)
	newCap := 2 * l
	if len(capacity) > 0 {
		newCap = capacity[0]
	}
	if c >= newCap {
		return buf
	}
	return grow(buf, newCap)
}

// Put puts back a byte buffer to the pool for reusing.
//
// The buf mustn't be touched after retuning it to the pool.
// Otherwise data races will occur.
func Put(buf *Buffer) {
	if !buf.noReuse {
		put(buf.B)
	}
	buf.B = nil
	buf.noReuse = false
	bpool.Put(buf)
}

// PutSlice puts back a byte slice which is obtained from function Grow.
//
// The byte slice mustn't be touched after returning it to the pool.
// Otherwise data races will occur.
func PutSlice(buf []byte) { put(buf) }

const (
	minPoolIdx = 3  // at least 8B (1<<3)
	poolSize   = 27 // max 64MB (1<<26)
)

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

// callers must guarantee that capacity is not greater than MaxBufSize.
func get(length, capacity int) []byte {
	idx := indexGet(capacity)
	out := sizedPools[idx].Get().([]byte)
	return out[:length]
}

func put(buf []byte) {
	cap_ := cap(buf)
	if shouldReuse(cap_) {
		idx := indexPut(cap_)
		buf = buf[:0]
		sizedPools[idx].Put(buf)
	}
}

func shouldReuse(cap int) bool {
	return cap >= MinBufSize && cap <= MaxBufSize
}

func grow(buf []byte, capacity int) []byte {
	len_, cap_ := len(buf), cap(buf)
	if double := 2 * cap_; capacity < double {
		capacity = double
	}

	var newBuf []byte
	if capacity > MaxBufSize {
		newBuf = make([]byte, len_, capacity)
	} else {
		newBuf = get(len_, capacity)
	}
	copy(newBuf, buf)
	put(buf)
	return newBuf
}

// indexGet finds the pool index for the given size to get buffer from,
// if size is not power of tow, it returns the next power of tow index.
func indexGet(size int) int {
	if size <= MinBufSize {
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

// isPowerOfTwo reports whether given integer is a power of two.
func isPowerOfTwo(n int) bool {
	return n&(n-1) == 0
}
