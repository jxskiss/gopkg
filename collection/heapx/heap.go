package heapx

import "unsafe"

// LessFunc is a comparator function to build a Heap.
type LessFunc[T any] func(lhs, rhs T) bool

// Heap implements the classic heap data-structure.
// A Heap is not safe for concurrent operations.
type Heap[T any] struct {
	items heapItems[T]
}

// NewHeap creates a new Heap.
func NewHeap[T any](cmp LessFunc[T]) *Heap[T] {
	h := &Heap[T]{}
	h.init(cmp)
	return h
}

func (h *Heap[T]) init(lessFunc LessFunc[T]) {
	h.items.init(lessFunc)
}

// Len returns the size of the heap.
func (h *Heap[T]) Len() int {
	return h.items.Len()
}

// Push pushes the element x onto the heap.
// The complexity is O(log n) where n = h.Len().
func (h *Heap[T]) Push(x T) {
	/*
		heap.Push(&h.items, x)
	*/
	h.items.Push(x)
	h.items.up(h.Len() - 1)
}

// Peek returns the minium element (according to the LessFunc) in the heap,
// it does not remove the item from the heap.
// The complexity is O(1).
func (h *Heap[T]) Peek() (x T, ok bool) {
	if h.items.Len() == 0 {
		return
	}
	return h.items.s0[0], true
}

// Pop removes and returns the minimum element (according to the LessFunc) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
func (h *Heap[T]) Pop() (x T, ok bool) {
	if h.items.Len() == 0 {
		return
	}

	/*
		return heap.Pop(&h.items).(T), true
	*/
	n := h.Len() - 1
	h.items.Swap(0, n)
	h.items.down(0, n)
	return h.items.Pop().(T), true
}

const (
	bktShift = 10
	bktSize  = 1 << bktShift
	bktMask  = bktSize - 1
	initSize = 64
	ptrSize  = unsafe.Sizeof(unsafe.Pointer(nil))

	shrinkThreshold = bktSize / 2
)

type heapItems[T any] struct {
	elemSz   uintptr
	lessFunc LessFunc[T]

	cap int
	len int
	s0  []T
	ss  []unsafe.Pointer
	ssp unsafe.Pointer // pointer to ss's data, i.e. &ss[0]
}

func (p *heapItems[T]) init(lessFunc LessFunc[T]) {
	p.elemSz = unsafe.Sizeof(*new(T))
	p.lessFunc = lessFunc
}

// index uses unsafe trick to eliminate slice bounds checking.
func (p *heapItems[T]) index(i int) *T {
	i, j := i>>bktShift, i&bktMask
	sptr := unsafe.Pointer(uintptr(p.ssp) + uintptr(i)*ptrSize)
	return (*T)(unsafe.Pointer(uintptr(*(*unsafe.Pointer)(sptr)) + uintptr(j)*p.elemSz))
}

func (p *heapItems[T]) addBucket(bkt []T) {
	p.ss = append(p.ss, unsafe.Pointer(&bkt[0]))
	p.ssp = unsafe.Pointer(&p.ss[0])
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
	if p.cap < p.len+1 {
		if p.len == 0 {
			p.s0 = make([]T, initSize)
			p.addBucket(p.s0)
			p.cap = initSize
		} else if p.cap < bktSize {
			newBkt := make([]T, p.cap*2)
			copy(newBkt, p.s0[:p.len])
			p.s0 = newBkt
			p.ss[0] = unsafe.Pointer(&newBkt[0])
			p.cap = len(newBkt)
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
	// Shrink buckets and free the underlying memory.
	if (p.len+shrinkThreshold)&bktMask == 0 && (p.cap-p.len) > bktSize {
		p.ss[len(p.ss)-1] = nil
		p.ss = p.ss[:len(p.ss)-1]
		p.cap -= bktSize
	}
	return ret
}

func (p *heapItems[T]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !p.lessFunc(*p.index(j), *p.index(i)) {
			break
		}
		p1, p2 := p.index(i), p.index(j)
		*p1, *p2 = *p2, *p1 // swap(i, j)
		j = i
	}
}

func (p *heapItems[T]) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && p.lessFunc(*p.index(j2), *p.index(j1)) {
			j = j2 // = 2*i + 2 // right child
		}
		if !p.lessFunc(*p.index(j), *p.index(i)) {
			break
		}
		p1, p2 := p.index(i), p.index(j)
		*p1, *p2 = *p2, *p1 // swap(i, j)
		i = j
	}
	return i > i0
}
