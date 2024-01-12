package heapx

import (
	"container/heap"
	"unsafe"
)

// LessFunc is a comparator function to build a Heap.
type LessFunc[T any] func(lhs, rhs T) bool

// Heap implements the classic heap data-structure.
// A Heap is not safe for concurrent operations.
type Heap[T any] struct {
	items heapItems[T]
}

// NewHeap creates a new Heap.
func NewHeap[T any](cmp LessFunc[T]) *Heap[T] {
	hp := &Heap[T]{}
	hp.init(cmp)
	return hp
}

func (p *Heap[T]) init(lessFunc LessFunc[T]) {
	p.items.elemSz = unsafe.Sizeof(*new(T))
	p.items.lessFunc = lessFunc
}

// Len returns the size of the heap.
func (p *Heap[T]) Len() int {
	return p.items.Len()
}

// Push adds an item to the heap.
func (p *Heap[T]) Push(x T) {
	heap.Push(&p.items, x)
}

// Peek returns the min item in the heap, according to the comparator
// function, it does not remove the item from the heap.
func (p *Heap[T]) Peek() (x T, ok bool) {
	if p.items.Len() == 0 {
		return
	}
	return p.items.s0[0], true
}

// Pop removes and returns the min item in the heap,
// according to the comparator function.
func (p *Heap[T]) Pop() (x T, ok bool) {
	if p.items.Len() == 0 {
		return
	}
	return heap.Pop(&p.items).(T), true
}

const (
	bktShift = 11
	bktSize  = 1 << bktShift
	bktMask  = bktSize - 1
	initSize = 64
	ptrSize  = unsafe.Sizeof(unsafe.Pointer(nil))
)

type heapItems[T any] struct {
	elemSz   uintptr
	lessFunc LessFunc[T]

	cap   int
	len   int
	s0    []T
	ss    []unsafe.Pointer
	ssPtr unsafe.Pointer
}

// index uses unsafe trick to eliminate slice bounds checking.
func (p *heapItems[T]) index(i int) *T {
	i, j := i>>bktShift, i&bktMask
	sPtr := unsafe.Pointer(uintptr(p.ssPtr) + uintptr(i)*ptrSize)
	return (*T)(unsafe.Pointer(uintptr(*(*unsafe.Pointer)(sPtr)) + uintptr(j)*p.elemSz))
}

func (p *heapItems[T]) addBucket(bkt []T) {
	p.ss = append(p.ss, unsafe.Pointer(&bkt[0]))
	p.ssPtr = unsafe.Pointer(&p.ss[0])
}

func (p *heapItems[T]) Len() int {
	return p.len
}

func (p *heapItems[T]) Less(i, j int) bool {
	return p.lessFunc(*p.index(i), *p.index(j))
}

func (p *heapItems[T]) Swap(i, j int) {
	p1, p2 := p.index(i), p.index(j)
	*p1, *p2 = *p2, *p1
}

func (p *heapItems[T]) Push(x any) {
	if p.len == 0 {
		p.s0 = make([]T, initSize)
		p.addBucket(p.s0)
		p.cap = initSize
	} else if p.cap < p.len+1 {
		if p.cap < bktSize {
			newBkt := make([]T, p.cap*2)
			copy(newBkt, p.s0[:p.len])
			p.s0 = newBkt
			p.ss[0] = unsafe.Pointer(&newBkt[0])
			p.cap *= 2
		} else {
			newBkt := make([]T, bktSize)
			p.addBucket(newBkt)
			p.cap += bktSize
		}
	}
	*p.index(p.len) = x.(T)
	p.len++
}

func (p *heapItems[T]) Pop() any {
	var ret, zero T
	if p.len > 0 {
		ptr := p.index(p.len - 1)
		ret, *ptr = *ptr, zero
		p.len--
	}
	// shrink buckets and free the underlying memory
	if p.len&bktMask == 0 && p.len > 0 {
		p.ss[len(p.ss)-1] = nil
		p.ss = p.ss[:len(p.ss)-1]
		p.cap -= bktSize
	}
	return ret
}
