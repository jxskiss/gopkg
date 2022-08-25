package bbp

import (
	"math/bits"
	"sync"
)

const (
	minShift = 6  // at least 64B
	maxShift = 25 // max 32MB

	// Min and max buffer size provided in this package.
	minBufSize = 1 << minShift
	maxBufSize = 1 << maxShift

	bigIdx     = 14 // 16KB
	minPoolIdx = 6
	maxPoolIdx = maxShift + (maxShift-bigIdx)*3
	poolSize   = maxPoolIdx + 1
)

// Get returns a byte slice from the pool with specified length and capacity.
// When you finish the work with the buffer, you may call Put to put it back
// to the pool for reusing.
func Get(length, capacity int) []byte {
	if capacity > maxBufSize {
		return make([]byte, length, capacity)
	}

	// Manually inlining.
	// return get(length, capacity)
	idx := indexGet(capacity)
	out := sizedPools[idx].Get().([]byte)
	return out[:length]
}

// Put puts back a byte slice to the pool for reusing.
//
// The byte slice mustn't be touched after returning it to the pool,
// otherwise data races will occur.
func Put(buf []byte) { put(buf) }

// Grow returns a new byte buffer from the pool which is at least of
// specified capacity.
//
// Note that if reuseBuf is true and a new slice is returned, the old
// buf will be put back to the pool, the caller must not retain reference
// to the buf and must not touch it after this calling returns, else
// data race happens.
func Grow(buf []byte, capacity int, reuseBuf bool) []byte {
	if cap(buf) >= capacity {
		return buf
	}
	return growWithOptions(buf, capacity, reuseBuf)
}

// -------- sized pools -------- //

var (
	powerOfTwoIdxTable = [maxShift + 1]int{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, // 0 - 13 [1B, 8KB]
		14, 18, 22, 26, 30, 34, 38, 42, 46, 50, 54, 58, // 14 - 26 [16KB, 32MB]
	}
	bufSizeTable [poolSize]int
	sizedPools   [poolSize]sync.Pool
)

func init() {
	for i := 0; i < poolSize; i++ {
		var size int
		if i < bigIdx {
			size = 1 << i
		} else {
			j, k := bigIdx+(i-bigIdx)/4, (i-bigIdx)%4
			quarter := (1 << j) / 4
			size = 1<<j + quarter*k
		}
		bufSizeTable[i] = size
		sizedPools[i].New = func() interface{} {
			buf := make([]byte, 0, size)
			return buf
		}
	}

	//for i := 0; i < poolSize; i++ {
	//	fmt.Printf("(%d) %d, ", i, bufSizeTable[i])
	//}
	//fmt.Println("")
}

// callers must guarantee that capacity is not greater than maxBufSize.
func get(length, capacity int) []byte {
	idx := indexGet(capacity)
	out := sizedPools[idx].Get().([]byte)
	return out[:length]
}

func put(buf []byte) {
	cap_ := cap(buf)
	if cap_ >= minBufSize && cap_ <= maxBufSize {
		idx := indexPut(cap_)
		buf = buf[:0]
		sizedPools[idx].Put(buf)
	}
}

func grow(buf []byte, capacity int) []byte {
	return growWithOptions(buf, capacity, true)
}

func growWithOptions(buf []byte, capacity int, reuseBuf bool) []byte {
	var newBuf []byte
	if capacity > maxBufSize {
		newBuf = make([]byte, len(buf), capacity)
	} else {
		// Manually inlining.
		// newBuf = get(len(buf), capacity)
		idx := indexGet(capacity)
		newBuf = sizedPools[idx].Get().([]byte)[:len(buf)]
	}
	copy(newBuf, buf)
	if reuseBuf {
		put(buf)
	}
	return newBuf
}

// indexGet finds the pool index for the given size to get buffer from,
// if size not equals to a predefined size, it returns the index of the
// next predefined size.
func indexGet(size int) int {
	if size <= minBufSize {
		return minPoolIdx
	}

	/*
		idx := bsr(size)
		if isPowerOfTow(size) {
		    return powerOfTwoIdxTable[idx]
		}
	*/

	// The following code is equivalent to the above commented lines,
	// they are manually inlined here to ensure that this function
	// will be inlined, for better performance.
	idx := bits.Len32(uint32(size)) - 1
	if size&(size-1) == 0 {
		return powerOfTwoIdxTable[idx]
	}

	if idx < bigIdx {
		return powerOfTwoIdxTable[idx] + 1
	}

	mod := size & (1<<idx - 1) // size % (2^idx)
	quarter := 1 << (idx - 2)  // (2^idx) / 4
	return powerOfTwoIdxTable[idx] + (mod+quarter-1)/quarter
}

// indexPut finds the pool index for the given size to put buffer back,
// if size not equals to a predefined size, it returns the index of the
// previous predefined size.
func indexPut(size int) int {
	idx := bsr(size)
	if isPowerOfTwo(size) || idx < bigIdx {
		return powerOfTwoIdxTable[idx]
	}
	mod := size & (1<<idx - 1) // size % (2^idx)
	quarter := 1 << (idx - 2)  // (2^idx) / 4
	return powerOfTwoIdxTable[idx] + mod/quarter
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
