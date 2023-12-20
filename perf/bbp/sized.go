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

	minPoolIdx = 6
	maxPoolIdx = maxShift
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
	return sizedPools[idx].Get(length)
}

// Put puts back a byte slice to the pool for reusing.
//
// The byte slice mustn't be touched after returning it to the pool,
// otherwise data races will occur.
func Put(buf []byte) { put(buf) }

// Grow checks capacity of buf, it returns a new byte buffer from the pool,
// if necessary, to guarantee space for another n bytes.
// After Grow(n), at least n bytes can be appended to the returned buffer
// without another allocation.
// If n is negative, Grow will panic.
//
// Note that if reuseBuf is true and a new slice is returned, the old
// buf will be put back to the pool, the caller must not retain reference
// to the old buf and must not access it again, else data race happens.
func Grow(buf []byte, n int, reuseBuf bool) []byte {
	if n < 0 {
		panic("bbp.Grow: negative size to grow")
	}
	if cap(buf) >= len(buf)+n {
		return buf
	}
	return growWithOptions(buf, len(buf)+n, reuseBuf)
}

// -------- sized pools -------- //

var (
	sizedPools [poolSize]*bufPool
)

func init() {
	for i := 0; i < poolSize; i++ {
		size := 1 << i
		sizedPools[i] = &bufPool{size: size}
	}
}

type bufPool struct {
	size int
	pool sync.Pool
}

func (p *bufPool) Get(length int) []byte {
	if ptr := p.pool.Get(); ptr != nil {
		return _toBuf(ptr.(unsafe.Pointer), length)
	}
	return make([]byte, length, p.size)
}

func (p *bufPool) Put(buf []byte) {
	if cap(buf) >= p.size {
		p.pool.Put(_toPtr(buf))
	}
}

func _toBuf(ptr unsafe.Pointer, length int) []byte {
	size := *(*int)(ptr)
	return *(*[]byte)(unsafe.Pointer(&unsafeheader.SliceHeader{
		Data: ptr,
		Len:  length,
		Cap:  size,
	}))
}

func _toPtr(buf []byte) unsafe.Pointer {
	h := *(*unsafeheader.SliceHeader)(unsafe.Pointer(&buf))
	*(*int)(h.Data) = h.Cap
	return h.Data
}

// callers must guarantee that capacity is not greater than maxBufSize.
func get(length, capacity int) []byte {
	idx := indexGet(capacity)
	return sizedPools[idx].Get(length)
}

func put(buf []byte) {
	c := cap(buf)
	if c >= minBufSize && c <= maxBufSize {
		idx := indexPut(c)
		ptr := _toPtr(buf)
		sizedPools[idx].pool.Put(ptr)
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
		newBuf = sizedPools[idx].Get(len(buf))
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

	// Manually inline bsr and isPowerOfTwo here.
	idx := bits.Len32(uint32(size))
	if size&(size-1) == 0 {
		idx -= 1 //nolint:revive
	}
	return idx
}

// indexPut finds the pool index for the given size to put buffer back,
// if size not equals to a predefined size, it returns the index of the
// previous predefined size.
func indexPut(size int) int {
	// Manually inline bsr.
	return bits.Len32(uint32(size)) - 1
}

// bsr.
//
// Callers within this package guarantee that n doesn't overflow int32.
//
//nolint:unused
func bsr(n int) int {
	return bits.Len32(uint32(n)) - 1
}

// isPowerOfTwo reports whether n is a power of two.
//
//nolint:unused
func isPowerOfTwo(n int) bool {
	return n&(n-1) == 0
}
