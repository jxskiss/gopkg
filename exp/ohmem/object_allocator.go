package ohmem

import (
	"reflect"
	"unsafe"
)

// NewObjectAllocator creates an object allocator.
// ohm will be used as the underlying memory pool.
func NewObjectAllocator[T any](ohm *OffHeapMem) *ObjectAllocator[T] {
	var x T
	var size = int(unsafe.Sizeof(x))
	var idx = ohm.getPoolIndex(size)
	return &ObjectAllocator[T]{
		ohm:  ohm,
		idx:  idx,
		size: size,
	}
}

// ObjectAllocator is an object pool to allocate and reuse memory of type T.
type ObjectAllocator[T any] struct {
	ohm  *OffHeapMem
	idx  int
	size int
}

// New returns a new object of type *T from the allocator.
func (a ObjectAllocator[T]) New() *T {
	buf := a.ohm.alloc(a.idx, a.size)
	h := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	return (*T)(unsafe.Pointer(h.Data))
}

// Free puts back an object to the allocator for reusing.
// If the memory is allocated from OS directly, if returns the memory to OS.
func (a ObjectAllocator[T]) Free(x *T) {
	h := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(x)),
	}
	bs := *(*[]byte)(unsafe.Pointer(&h))
	a.ohm.Free(bs)
}
