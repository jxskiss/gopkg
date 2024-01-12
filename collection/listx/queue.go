//nolint:dupl
package listx

// Queue is a First In First Out data structure implementation.
// The zero value for Queue is an empty queue ready to use.
// A Queue is not safe for concurrent operations.
type Queue[T any] struct {
	l List[T]
}

// NewQueue creates a new Queue instance.
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

// Len returns the size of the Queue.
func (q *Queue[T]) Len() int {
	return q.l.Len()
}

// Enqueue adds an item at the back of the Queue in *O(1)* time complexity.
func (q *Queue[T]) Enqueue(item T) {
	q.l.PushFront(item)
}

// Dequeue removes and returns the Queue's front item in *O(1)* time complexity.
func (q *Queue[T]) Dequeue() (item T, ok bool) {
	last := q.l.Back()
	if last != nil {
		item = q.l.Remove(last)
		ok = true
	}
	return
}

// Peek returns the Queue's front item in *O(1)* time complexity,
// it does not remove the item from the Queue.
func (q *Queue[T]) Peek() (item T, ok bool) {
	last := q.l.Back()
	if last != nil {
		item, ok = last.Value.(T), true
	}
	return
}
