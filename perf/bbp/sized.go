package bbp

import (
	"math/bits"
	"sync"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

const (
	minShift = 6  // at least 64B
	maxShift = 25 // max 32MB

	// Min and max buffer size provided in this package.
	minBufSize = 1 << minShift
	maxBufSize = 1 << maxShift

	idx4KB   = 12       // 4KB
	idx8KB   = 13       // 8KB
	idx12KB  = 14       // 8KB + 4KB
	idx16KB  = 15       // 8KB + (2 * 4KB)
	size4KB  = 4 << 10  // 4KB
	size8KB  = 8 << 10  // 8KB
	size12KB = 12 << 10 // 8KB + 4KB
	size16KB = 16 << 10 // 8KB + (2 * 4KB)

	minPoolIdx = 6
	maxPoolIdx = maxShift + (maxShift-idx8KB)*3 - 2
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
	ptr := sizedPools[idx].Get().(unsafe.Pointer)
	return _toBuf(ptr, length)
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
		15, 19, 23, 27, 31, 35, 39, 43, 47, 51, 55, 59, // 14 - 26 [16KB, 32MB]
	}
	bufSizeTable [poolSize]int
	sizedPools   [poolSize]sync.Pool
)

func init() {
	for i := 0; i < poolSize; i++ {
		var size int
		if i <= 13 { // <= 8KB (idx8KB)
			size = 1 << i
		} else if i == 14 {
			size = size12KB // 12KB
		} else { // i >= 15, size >= 16KB
			j, k := 14+(i-15)/4, (i-15)%4
			quarter := (1 << j) / 4
			size = 1<<j + quarter*k
		}
		bufSizeTable[i] = size
		sizedPools[i].New = func() any {
			buf := make([]byte, 0, size)
			return _toPtr(buf)
		}
	}

	//for i := 0; i < poolSize; i++ {
	//	fmt.Printf("(%d) %d, ", i, bufSizeTable[i])
	//}
	//fmt.Println("")
}

func _toBuf(ptr unsafe.Pointer, length int) []byte {
	size := *(*int)(ptr)
	return *(*[]byte)(unsafe.Pointer(&unsafeheader.Slice{
		Data: ptr,
		Len:  length,
		Cap:  size,
	}))
}

func _toPtr(buf []byte) unsafe.Pointer {
	h := *(*unsafeheader.Slice)(unsafe.Pointer(&buf))
	*(*int)(h.Data) = h.Cap
	return h.Data
}

// callers must guarantee that capacity is not greater than maxBufSize.
func get(length, capacity int) []byte {
	idx := indexGet(capacity)
	ptr := sizedPools[idx].Get().(unsafe.Pointer)
	return _toBuf(ptr, length)
}

func put(buf []byte) {
	c := cap(buf)
	if c >= minBufSize && c <= maxBufSize {
		idx := indexPut(c)
		ptr := _toPtr(buf)
		sizedPools[idx].Put(ptr)
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
		ptr := sizedPools[idx].Get().(unsafe.Pointer)
		newBuf = _toBuf(ptr, len(buf))
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
//
// See indexGet_readable for a more readable version of this function.
func indexGet(size int) int {
	if size <= minBufSize {
		return minPoolIdx
	}

	// For better performance, bsr and isPowerOfTwo are inlined here
	// to ensure that this function can be inlined.
	p2i := bits.Len32(uint32(size)) - 1
	idx := powerOfTwoIdxTable[p2i]

	if size&(size-1) != 0 { // not power of two
		if size > size16KB {
			mod := size & (1<<p2i - 1)    // size % (2^p2i)
			idx += (mod - 1) >> (p2i - 2) // (mod - 1) / (2^p2i / 4)
		} else if size > size12KB {
			idx = idx12KB
		}
		idx++
	}
	return idx
}

//nolint:all
func indexGet_readable(size int) int {
	if size <= minBufSize {
		return minPoolIdx
	}

	idx := bsr(size)
	if isPowerOfTwo(size) {
		return powerOfTwoIdxTable[idx]
	}

	if idx < idx8KB {
		return powerOfTwoIdxTable[idx] + 1
	}

	if idx == idx8KB { // (8KB, 16KB)
		const half = size4KB
		mod := size & (1<<idx - 1) // size % (2^idx)
		return powerOfTwoIdxTable[idx] + (mod+half-1)/half
	}

	// larger than 16KB
	mod := size & (1<<idx - 1) // size % (2^idx)
	quarter := 1 << (idx - 2)  // (2^idx) / 4
	return powerOfTwoIdxTable[idx] + (mod+quarter-1)/quarter
}

// indexPut finds the pool index for the given size to put buffer back,
// if size not equals to a predefined size, it returns the index of the
// previous predefined size.
//
// See indexPut_readable for a more readable version of this function.
func indexPut(size int) int {
	// For better performance, bsr and isPowerOfTwo are inlined here
	// to ensure that this function can be inlined.
	p2i := bits.Len32(uint32(size)) - 1
	idx := powerOfTwoIdxTable[p2i]

	if size&(size-1) != 0 { // not power of two
		if size > size16KB {
			mod := size & (1<<p2i - 1) // size % (2^p2i)
			idx += mod >> (p2i - 2)    // mod / (2^p2i / 4)
		} else if size >= size12KB {
			idx = idx12KB
		}
	}
	return idx
}

//nolint:all
func indexPut_readable(size int) int {
	idx := bsr(size)
	if isPowerOfTwo(size) || idx < idx8KB {
		return powerOfTwoIdxTable[idx]
	}

	if idx == idx8KB {
		const half = size4KB
		mod := size & (1<<idx - 1) // size % (2^idx)
		return powerOfTwoIdxTable[idx] + mod/half
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
