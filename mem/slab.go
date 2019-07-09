package mem

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// Slab allocates memory in batch size and provides small chunks by the
// public APIs, it's designed to be fast and reduce GC pressure in case of
// allocating many temporary small bytes or int(N) or uint(N) slice.
//
// Don't use this to allocate long-living objects, the underlying heap
// memory won't be freed until all allocated objects been released.
//
// It uses unsafe tricks but is very fast and safe for concurrent use.
// The zero value of Slab allocates no memory and is ready to use.
type Slab struct {
	Size      int
	Threshold int

	mu  sync.Mutex
	buf unsafe.Pointer
}

// DefaultSlab is meant to be used by multiple packages to share the
// underlying memory block, it allocates no memory if it's never used.
var DefaultSlab = &Slab{Size: 1 << 20} // 1MB

const defaultThreshold = 2048      // 2KB
const defaultBlockSize = 256 << 10 // 256KB
type block struct {
	p int64
	b []byte
}

func (m *Slab) size() int64 {
	if m.Size > 0 {
		return int64(m.Size)
	}
	return defaultBlockSize
}

func (m *Slab) threshold() int {
	if m.Threshold > 0 {
		return m.Threshold
	}
	return defaultThreshold
}

func (m *Slab) newblock() *block {
	return &block{b: make([]byte, m.size())}
}

func (m *Slab) block(p *unsafe.Pointer) *block {
	return (*block)(atomic.LoadPointer(p))
}

func (m *Slab) Bytes(size int) []byte {
	// allocate big chunk directly from heap
	if size >= m.threshold() {
		return make([]byte, 0, size)
	}

	x := int64(size)
	b := m.block(&m.buf)
	if b == nil {
		m.mu.Lock()
		if b = m.block(&m.buf); b == nil {
			b = m.newblock()
			atomic.StorePointer(&m.buf, unsafe.Pointer(b))
		}
		m.mu.Unlock()
	}
	i := atomic.AddInt64(&b.p, x)
	for i > int64(m.size()) {
		m.mu.Lock()
		old := b
		if b = m.block(&m.buf); b == old {
			b = m.newblock()
			atomic.StorePointer(&m.buf, unsafe.Pointer(b))
		}
		i = atomic.AddInt64(&b.p, x)
		m.mu.Unlock()
	}
	return b.b[i-x : i-x : i]
}

func (m *Slab) Int8(size int) []int8 {
	s := m.Bytes(size)
	return *(*[]int8)(unsafe.Pointer(&s))
}

func (m *Slab) Uint8(size int) []uint8 {
	s := m.Bytes(size)
	return *(*[]uint8)(unsafe.Pointer(&s))
}

func (m *Slab) Int16(size int) []int16 {
	s := m.Bytes(size * 2)
	h := (*sliceHeader)(unsafe.Pointer(&s))
	h.cap /= 2
	return *(*[]int16)(unsafe.Pointer(h))
}

func (m *Slab) Uint16(size int) []uint16 {
	s := m.Bytes(size * 2)
	h := (*sliceHeader)(unsafe.Pointer(&s))
	h.cap /= 2
	return *(*[]uint16)(unsafe.Pointer(h))
}

func (m *Slab) Int32(size int) []int32 {
	s := m.Bytes(size * 4)
	h := (*sliceHeader)(unsafe.Pointer(&s))
	h.cap /= 4
	return *(*[]int32)(unsafe.Pointer(h))
}

func (m *Slab) Uint32(size int) []uint32 {
	s := m.Bytes(size * 4)
	h := (*sliceHeader)(unsafe.Pointer(&s))
	h.cap /= 4
	return *(*[]uint32)(unsafe.Pointer(h))
}

func (m *Slab) Int64(size int) []int64 {
	s := m.Bytes(size * 8)
	h := (*sliceHeader)(unsafe.Pointer(&s))
	h.cap /= 8
	return *(*[]int64)(unsafe.Pointer(h))
}

func (m *Slab) Uint64(size int) []uint64 {
	s := m.Bytes(size * 8)
	h := (*sliceHeader)(unsafe.Pointer(&s))
	h.cap /= 8
	return *(*[]uint64)(unsafe.Pointer(h))
}

type sliceHeader struct {
	ptr unsafe.Pointer
	len int
	cap int
}
