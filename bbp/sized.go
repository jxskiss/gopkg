package bbp

import (
	"math/bits"
	"sync"
)

const (
	// MinBufSize is the minimum buffer size (8B) provided in this package.
	MinBufSize = 8

	// MaxBufSize is the maximum buffer size (64MB) provided in this package.
	MaxBufSize = 1 << 26

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

func get(length int, capacity ...int) []byte {
	cap_ := length
	if len(capacity) > 0 && capacity[0] > length {
		cap_ = capacity[0]
	}
	if cap_ > MaxBufSize {
		return make([]byte, length, cap_)
	}
	idx := indexGet(cap_)
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
		idx := indexGet(capacity)
		newBuf = sizedPools[idx].Get().([]byte)[:len_]
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
