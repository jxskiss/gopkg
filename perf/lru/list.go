package lru

import "unsafe"

type element struct {
	next, prev uint32

	key, value interface{}

	expires int64 // nanosecond timestamp
	index   uint32
}

func newList(capacity int) *list {
	elems := make([]element, capacity+1)
	l := &list{
		elems: elems,
		root:  &elems[0],
	}

	size := len(elems)
	for i := 1; i < size; i++ {
		e := &elems[i]
		e.index = uint32(i)
		l.PushBack(e)
	}
	return l
}

type list struct {
	elems []element
	root  *element
	len   int
}

func (l *list) Front() *element {
	if l.len == 0 {
		return nil
	}
	return l.get(l.root.next)
}

func (l *list) Back() *element {
	if l.len == 0 {
		return nil
	}
	return l.get(l.root.prev)
}

func (l *list) PushFront(elem *element) *element {
	return l.insert(elem, l.root)
}

func (l *list) PushBack(elem *element) *element {
	return l.insert(elem, l.get(l.root.prev))
}

func (l *list) MoveToFront(elem *element) {
	l.insert(l.remove(elem), l.root)
}

func (l *list) MoveToBack(elem *element) {
	l.insert(l.remove(elem), l.get(l.root.prev))
}

func (l *list) insert(elem, at *element) *element {
	next := l.get(at.next)
	at.next = elem.index
	elem.prev = at.index
	elem.next = next.index
	next.prev = elem.index
	l.len++
	return elem
}

func (l *list) remove(elem *element) *element {
	prev := l.get(elem.prev)
	next := l.get(elem.next)
	prev.next = elem.next
	next.prev = elem.prev
	l.len--
	return elem
}

const elemsize = unsafe.Sizeof(element{})

// get uses unsafe trick to eliminate slice bounds checking
func (l *list) get(idx uint32) *element {
	return (*element)(unsafe.Pointer(uintptr(unsafe.Pointer(l.root)) + uintptr(idx)*elemsize))
}
