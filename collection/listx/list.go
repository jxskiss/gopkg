package listx

import (
	"container/list"
	"unsafe"
)

func init() {
	var l List[string]
	if unsafe.Pointer(&l) != unsafe.Pointer(&l.l) {
		panic("shit happens")
	}
}

// Element is an element of a linked list.
type Element = list.Element

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type List[T any] struct {
	l list.List
}

// NewList returns an initialized list.
func NewList[T any]() *List[T] {
	return (*List[T])(unsafe.Pointer(list.New()))
}

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *List[T]) Len() int {
	return l.l.Len()
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List[T]) Front() *Element {
	return l.l.Front()
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[T]) Back() *Element {
	return l.l.Back()
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *List[T]) Remove(e *Element) T {
	return (*list.List)(unsafe.Pointer(l)).Remove(e).(T)
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List[T]) PushFront(v T) *Element {
	return (*list.List)(unsafe.Pointer(l)).PushFront(v)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[T]) PushBack(v T) *Element {
	return (*list.List)(unsafe.Pointer(l)).PushBack(v)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertBefore(v T, mark *Element) *Element {
	return (*list.List)(unsafe.Pointer(l)).InsertBefore(v, mark)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertAfter(v T, mark *Element) *Element {
	return (*list.List)(unsafe.Pointer(l)).InsertAfter(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToFront(e *Element) {
	l.l.MoveToFront(e)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToBack(e *Element) {
	l.l.MoveToBack(e)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveBefore(e, mark *Element) {
	l.l.MoveBefore(e, mark)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveAfter(e, mark *Element) {
	l.l.MoveAfter(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushBackList(other *List[T]) {
	l.l.PushBackList((*list.List)(unsafe.Pointer(other)))
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushFrontList(other *List[T]) {
	l.l.PushFrontList((*list.List)(unsafe.Pointer(other)))
}
