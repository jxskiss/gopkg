package heapx

type PriorityQueue[P Ordered, V any] struct {
	heap Heap[pqItem[P, V]]
}

type pqItem[P Ordered, V any] struct {
	priority P
	value    V
}

func (x pqItem[P, V]) Compare(other pqItem[P, V]) int {
	if x.priority < other.priority {
		return -1
	}
	if x.priority > other.priority {
		return 1
	}
	return 0
}

func NewPriorityQueue[P Ordered, V any]() *PriorityQueue[P, V] {
	return &PriorityQueue[P, V]{}
}

func (pq *PriorityQueue[P, V]) Len() int {
	return pq.heap.Len()
}

func (pq *PriorityQueue[P, V]) Push(value V, priority P) {
	pq.heap.Push(pqItem[P, V]{
		priority: priority,
		value:    value,
	})
}

func (pq *PriorityQueue[P, V]) Pop() (value V, priority P, ok bool) {
	item, ok := pq.heap.Pop()
	if ok {
		value, priority = item.value, item.priority
	}
	return
}

func (pq *PriorityQueue[P, V]) Peek() (value V, priority P, ok bool) {
	item, ok := pq.heap.Peek()
	if ok {
		value, priority = item.value, item.priority
	}
	return
}
