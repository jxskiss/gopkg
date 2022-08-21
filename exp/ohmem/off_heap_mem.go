// Package ohmem manages off-heap memory.
//
// This package is a fork of
// https://gist.github.com/Petelin/9421c156893e7e72e33d41ea60fcd60f
//
// Also see this blog post:
// https://dgraph.io/blog/post/manual-memory-management-golang-jemalloc/
//
// This package provides utilities to manually allocate and free memory
// from the operating system, it's the users' responsibility to take care
// of memory safety.
//
// By default, this package uses mmap to allocate memory from OS,
// optionally user can change to use mi-malloc (cgo) by specifying
// a build tag "mimalloc".
//
// Warning: note that this package was written for fun, it's experimental
// and the code is not well-tested.
// Generally you don't need this package, and it's hard to correctly
// manage memory manually in Go.
//
// Warning: also note that with the runtime allocator, allocating directly
// from OS is PAGE_SIZE aligned, you may waste a lot of memory if you are
// allocating many small memory blocks.
//
// Warning: code in this package is not well-tested.
package ohmem

import (
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"
)

//go:linkname throw runtime.throw
func throw(s string)

// OffHeapMem manages some pre-allocated memory in free lists.
// An OffHeapMem instance should be reused in the whole lifetime of a program.
//
// When user calls Alloc, it checks for a block of free memory to use
// if available, else it allocates memory from OS directly.
//
// When user calls Free, if the memory is pre-allocated it puts it
// back to free lists for further reusing, else it returns the memory
// to the OS.
type OffHeapMem struct {
	classes []Class
	maxSize int

	mem   []byte
	pools []_pool
	base  uintptr
	max   uintptr

	released uintptr
}

type Class struct {
	Size int
	Cap  int
}

type _pool struct {
	*sync.Mutex
	ptrs *[]uintptr
	len  *int
	base uintptr
	max  uintptr
}

// NewOffHeapMem creates a new OffHeapMem instance using given classes.
// If the OffHeapMem instance is not needed anymore, user should call
// *OffHeapMem.Release to free the memory.
func NewOffHeapMem(classes []Class) *OffHeapMem {
	var maxSize int
	classes = tidyClasses(classes)
	maxSize = classes[0].Size

	pools := make([]_pool, len(classes))
	sumCap := 0
	for i := 0; i < len(pools); i++ {
		class := classes[i]
		sumCap += class.Size * class.Cap
	}

	// 在 runtime 管理的堆之外分配内存，GC 对这部分内存是忽略的
	mem := _C_zalloc(sumCap)

	h := (*reflect.SliceHeader)(unsafe.Pointer(&mem))
	baseAddr := h.Data
	for i := 0; i < len(pools); i++ {
		var ptrs []uintptr
		class := classes[i]
		for j := 0; j < class.Cap; j++ {
			ptrs = append(ptrs, baseAddr)
			baseAddr += uintptr(class.Size)
		}
		length := class.Cap
		pools[i] = _pool{
			Mutex: new(sync.Mutex),
			ptrs:  &ptrs,
			len:   &length,
			base:  ptrs[0],
			max:   ptrs[len(ptrs)-1],
		}
	}

	oohMem := &OffHeapMem{
		classes: classes,
		maxSize: maxSize,
		mem:     mem,
		pools:   pools,
		base:    pools[0].base,
		max:     pools[len(pools)-1].max,
	}
	runtime.SetFinalizer(oohMem, (*OffHeapMem).Release)
	return oohMem
}

// class.Size 按从大到小排序，Alloc 函数依赖这个排序
// 如果 class.Size 重复，合并 class.Cap
func tidyClasses(classes []Class) []Class {

	// 传入的 class 可能不是对齐到 64 字节，这里调整为每个对象都对齐
	for i := 0; i < len(classes); i++ {
		classes[i].Size = align64(classes[i].Size)
	}

	sort.Slice(classes, func(i, j int) bool {
		return classes[j].Size < classes[i].Size
	})
	var i = 0
	for j := 1; j < len(classes); j++ {
		if classes[j].Size == classes[i].Size {
			classes[i].Cap += classes[j].Cap
			continue
		}
		i++
		classes[i] = classes[j]
	}
	classes = classes[:i+1]
	return classes
}

func align64(n int) int {
	return (n + 63) / 64 * 64
}

func (a *OffHeapMem) isInitialized() bool {
	return a != nil && len(a.mem) > 0
}

func (a *OffHeapMem) getPoolIndex(size int) int {
	if !a.isInitialized() {
		return -1
	}
	// 使用最小的合适的 class，二分查找
	idx, j := 0, len(a.classes)
	for idx < j {
		// idx ≤ h < j
		h := (idx + j) >> 1
		if a.classes[h].Size < size {
			j = h
		} else {
			idx = h + 1
		}
	}
	if idx == len(a.classes) || a.classes[idx].Size < size {
		idx -= 1
	}
	return idx
}

// Release returns the pre-allocated memory back to OS.
//
// It's the user's responsibility to make sure that all memory
// allocated from the instance have been freed, and they won't be
// accessed after it's released, else undefined behavior happens.
func (a *OffHeapMem) Release() {
	if a.isInitialized() {
		if atomic.CompareAndSwapUintptr(&a.released, 0, 1) {
			_C_free(a.mem)
			a.mem = nil
			a.pools = nil
		}
		runtime.SetFinalizer(a, nil)
	}
}

// Alloc allocates memory from OffHeapMem.
func (a *OffHeapMem) Alloc(size int) []byte {
	return a.alloc(-1, size)
}

func (a *OffHeapMem) alloc(idx int, size int) []byte {
	if !a.isInitialized() || size > a.maxSize {
		return _C_zalloc(size)
	}

	buf, ok := a.tryAllocFromPool(idx, size)
	if ok {
		return buf
	}

	// 内存池没有初始化，或者没有空闲内存，向操作系统申请分配
	aligned := align64(size)
	buf = _C_zalloc(aligned)
	return buf[:size:size]
}

func (a *OffHeapMem) tryAllocFromPool(idx int, size int) ([]byte, bool) {
	if a.isInitialized() {
		if idx < 0 {
			idx = a.getPoolIndex(size)
		}
		pool := a.pools[idx]
		pool.Lock()
		if *pool.len > 0 {
			ptr := (*pool.ptrs)[*pool.len-1]
			*pool.len -= 1
			pool.Unlock()

			bs := _getBytes(ptr, size, size)
			_wipe(bs)
			return bs, true
		}
		pool.Unlock()
	}
	return nil, false
}

// Free returns memory to OffHeapMem for reusing, if the memory
// is allocated from OS directly, it returns the memory to OS.
func (a *OffHeapMem) Free(bs []byte) {
	if !a.isInitialized() || cap(bs) > a.maxSize {
		_C_free(bs)
		return
	}
	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&bs)).Data
	if ptr < a.base || ptr > a.max {
		_C_free(bs)
		return
	}

	// pools 是有序的，二分查找
	pools := a.pools
	for len(pools) > 0 {
		i := len(pools) / 2
		pool := pools[i]
		if ptr < pool.base {
			pools = pools[:i]
			continue
		} else if ptr > pool.max {
			pools = pools[i:]
			continue
		}

		pool.Lock()
		(*pool.ptrs)[*pool.len] = ptr
		*pool.len += 1
		pool.Unlock()
		return
	}
}

func _getBytes(ptr uintptr, len, cap int) []byte {
	var sl = reflect.SliceHeader{Data: ptr, Len: len, Cap: cap}
	return *(*[]byte)(unsafe.Pointer(&sl))
}

func _wipe(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
	runtime.KeepAlive(buf)
}
