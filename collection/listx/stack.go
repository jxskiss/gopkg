//nolint:dupl
package listx

// Stack implements a Last In First Out data structure built upon
// a doubly-linked list.
// The zero value for Stack is an empty stack ready to use.
// A Stack is not safe for concurrent operations.
//
// In most cases, a simple slice based implementation is a better choice.
type Stack[T any] struct {
	l List[T]
}

// NewStack creates a new Stack instance.
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

// Len returns the size of the Stack.
func (s *Stack[T]) Len() int {
	return s.l.Len()
}

// Push adds on an item on the top of the Stack in *O(1)* time complexity.
func (s *Stack[T]) Push(item T) {
	s.l.PushFront(item)
}

// Pop removes and returns the item on the top of the Stack in *O(1)* time complexity.
func (s *Stack[T]) Pop() (item T, ok bool) {
	first := s.l.Front()
	if first != nil {
		item = s.l.Remove(first)
		ok = true
	}
	return
}

// Peek returns the item on the top of the Stack in *O(1)* time complexity,
// it does not remove the item from the Stack.
func (s *Stack[T]) Peek() (item T, ok bool) {
	first := s.l.Front()
	if first != nil {
		item, ok = first.Value.(T), true
	}
	return
}
