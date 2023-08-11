package heapx

import "container/heap"

type Cmp[T any] interface {
	Compare(other T) int
}

type cmpItems[T Cmp[T]] []T

func (p cmpItems[T]) Len() int {
	return len(p)
}

func (p cmpItems[T]) Less(i, j int) bool {
	return p[i].Compare(p[j]) < 0
}

func (p cmpItems[T]) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *cmpItems[T]) Push(x any) {
	*p = append(*p, x.(T))
}

func (p *cmpItems[T]) Pop() any {
	var zero T
	old := *p
	n := len(old)
	item := old[n-1]
	old[n-1] = zero // avoid memory leak
	*p = old[:n-1]
	return item
}

type Heap[T Cmp[T]] struct {
	items cmpItems[T]
}

func NewHeap[T Cmp[T]]() *Heap[T] {
	return &Heap[T]{}
}

func (p *Heap[T]) Len() int {
	return p.items.Len()
}

func (p *Heap[T]) Push(x T) {
	heap.Push(&p.items, x)
}

func (p *Heap[T]) Peek() (x T, ok bool) {
	if p.items.Len() == 0 {
		return x, false
	}
	return p.items[0], true
}

func (p *Heap[T]) Pop() (x T, ok bool) {
	if p.items.Len() == 0 {
		var zero T
		return zero, false
	}
	return heap.Pop(&p.items).(T), true
}
