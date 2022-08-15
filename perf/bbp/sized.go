package bbp

import (
	"math/bits"
	"sync"
)

const (
	minPoolIdx = 6  // at least 1<<6 = 64B
	maxPoolIdx = 25 // max 1<<25 = 32MB

	poolSize = maxPoolIdx + 1
)

const (
	// minSize is the minimum buffer size provided in this package.
	minSize = 1 << minPoolIdx

	// maxSize is the maximum buffer size provided in this package.
	maxSize = 1 << maxPoolIdx
)

// Get returns a byte slice from the pool with specified length and capacity.
// When you finish the work with the buffer, you may call Put to put it back
// to the pool for reusing.
func Get(length, capacity int) []byte {
	if capacity > maxSize {
		return make([]byte, length, capacity)
	}
	return get(length, capacity)
}

// Put puts back a byte slice to the pool for reusing.
//
// The byte slice mustn't be touched after returning it to the pool,
// otherwise data races will occur.
func Put(buf []byte) { put(buf) }

// Grow returns a new byte buffer from the pool which is at least of
// specified capacity.
//
// Note that if a new slice is returned, the old one will be put back
// to the pool, it this is not desired, you should use Get and Put
// to take full control of the lifetime of buf.
func Grow(buf []byte, capacity int) []byte {
	if cap(buf) >= capacity {
		return buf
	}
	return grow(buf, capacity)
}

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

func grow(buf []byte, capacity int) []byte {
	var newBuf []byte
	if capacity > maxSize {
		newBuf = make([]byte, len(buf), capacity)
	} else {
		newBuf = get(len(buf), capacity)
	}
	copy(newBuf, buf)
	put(buf)
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
