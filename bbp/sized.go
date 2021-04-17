package bbp

import (
	"math/bits"
	"sync"
)

const (
	minPoolIdx = 3       // at least 8 bytes
	minBufSize = 8       // at least 8 bytes
	maxBufSize = 1 << 25 // max 32 MB
	poolSize   = 26      // max 32 MB
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
	if cap_ > maxBufSize {
		return make([]byte, length, cap_)
	}
	idx := index(cap_)
	out := sizedPools[idx].Get().([]byte)
	return out[:length]
}

func put(buf []byte) {
	cap_ := cap(buf)
	if canReuse(cap_) {
		idx := bsr(cap_)
		buf = buf[:0]
		sizedPools[idx].Put(buf)
	}
}

func canReuse(cap int) bool {
	return cap >= minBufSize && cap <= maxBufSize && isPowerOfTwo(cap)
}

func grow(buf []byte, capacity ...int) []byte {
	len_, cap_ := len(buf), cap(buf)
	newCap := cap_ * 2
	if len(capacity) > 0 && capacity[0] > newCap {
		newCap = capacity[0]
	}

	if newCap > maxBufSize {
		newBuf := make([]byte, len_, newCap)
		copy(newBuf, buf)
		put(buf)
		return newBuf
	}

	newCap = ceilToPowerOfTwo(newCap)
	idx := index(newCap)
	newBuf := sizedPools[idx].Get().([]byte)
	copy(newBuf, buf)
	put(buf)
	return newBuf
}

// index finds the pool index for the given size, if size is not
// power of tow, it returns the next power of tow pool index.
func index(size int) int {
	if size <= minBufSize {
		return minPoolIdx
	}
	idx := bsr(size)
	if isPowerOfTwo(size) {
		return idx
	}
	return idx + 1
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

// ceilToPowerOfTwo returns the least power of two integer value greater than
// or equal to n.
// Callers within this package guarantee that n doesn't overflow int32.
func ceilToPowerOfTwo(n int) int {
	if n <= 2 {
		return n
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	// n |= n >> 32
	n++
	return n
}
