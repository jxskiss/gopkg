package heapx

import "github.com/jxskiss/gopkg/v2/internal/constraints"

type Ordered = constraints.Ordered

// PriorityQueue is a heap-based priority queue implementation.
//
// It can be either min (ascending) or max (descending) oriented/ordered.
// The type parameters `P` and `V` specify the type of the underlying
// priority and value.
//
// A PriorityQueue is not safe for concurrent operations.
type PriorityQueue[P Ordered, V any] struct {
	heap Heap[pqItem[P, V]]
}

type pqItem[P Ordered, V any] struct {
	priority P
	value    V
}

// NewMaxPriorityQueue creates a new maximum oriented PriorityQueue.
func NewMaxPriorityQueue[P Ordered, V any]() *PriorityQueue[P, V] {
	pq := &PriorityQueue[P, V]{}
	pq.heap.init(func(lhs, rhs pqItem[P, V]) bool {
		return rhs.priority < lhs.priority
	})
	return pq
}

// NewMinPriorityQueue creates a new minimum oriented PriorityQueue.
func NewMinPriorityQueue[P Ordered, V any]() *PriorityQueue[P, V] {
	pq := &PriorityQueue[P, V]{}
	pq.heap.init(func(lhs, rhs pqItem[P, V]) bool {
		return lhs.priority < rhs.priority
	})
	return pq
}

// Len returns the size of the PriorityQueue.
func (pq *PriorityQueue[P, V]) Len() int {
	return pq.heap.Len()
}

// Push adds a value with priority to the PriorityQueue.
func (pq *PriorityQueue[P, V]) Push(priority P, value V) {
	pq.heap.Push(pqItem[P, V]{
		priority: priority,
		value:    value,
	})
}

// Peek returns the most priority value in the PriorityQueue,
// it does not remove the value from the queue.
func (pq *PriorityQueue[P, V]) Peek() (priority P, value V, ok bool) {
	item, ok := pq.heap.Peek()
	if ok {
		priority, value = item.priority, item.value
	}
	return
}

// Pop removes and returns the most priority value in the PriorityQueue.
func (pq *PriorityQueue[P, V]) Pop() (priority P, value V, ok bool) {
	item, ok := pq.heap.Pop()
	if ok {
		priority, value = item.priority, item.value
	}
	return
}
