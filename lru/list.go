package lru

import "time"

type element struct {
	next, prev *element

	key, value interface{}

	expires int64 // nanosecond timestamp
	index   uint32
}

func expires(ttl time.Duration) (expires int64) {
	return time.Now().Add(ttl).UnixNano()
}

func isExpired(expires int64) bool {
	return expires > 0 && expires < time.Now().UnixNano()
}

func newList(elems []element) *list {
	l := &list{}
	l.root.next = &l.root
	l.root.prev = &l.root

	size := len(elems)
	for i := 0; i < size; i++ {
		e := &elems[i]
		e.index = uint32(i)
		l.PushBack(e)
	}
	return l
}

type list struct {
	root element
	len  int
}

func (l *list) Front() *element {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

func (l *list) Back() *element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *list) PushFront(elem *element) *element {
	return l.insert(elem, &l.root)
}

func (l *list) PushBack(elem *element) *element {
	return l.insert(elem, l.root.prev)
}

func (l *list) MoveToFront(elem *element) {
	l.insert(l.remove(elem), &l.root)
}

func (l *list) MoveToBack(elem *element) {
	l.insert(l.remove(elem), l.root.prev)
}

func (l *list) insert(elem, at *element) *element {
	next := at.next
	at.next = elem
	elem.prev = at
	elem.next = next
	next.prev = elem
	l.len++
	return elem
}

func (l *list) remove(elem *element) *element {
	elem.prev.next = elem.next
	elem.next.prev = elem.prev
	//elem.next = nil
	//elem.prev = nil
	l.len--
	return elem
}
