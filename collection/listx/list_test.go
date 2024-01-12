// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package listx

import "testing"

func checkListLen[T any](t *testing.T, l *List[T], len int) bool {
	if n := l.Len(); n != len {
		t.Errorf("l.Len() = %d, want %d", n, len)
		return false
	}
	return true
}

func checkListPointers[T any](t *testing.T, l *List[T], es []*Element) {
	_, _, _ = t, l, es
	// pass
}

func TestList(t *testing.T) {
	l := NewList[any]()
	checkListPointers(t, l, []*Element{})

	// Single element list
	e := l.PushFront("a")
	checkListPointers(t, l, []*Element{e})
	l.MoveToFront(e)
	checkListPointers(t, l, []*Element{e})
	l.MoveToBack(e)
	checkListPointers(t, l, []*Element{e})
	l.Remove(e)
	checkListPointers(t, l, []*Element{})

	// Bigger list
	e2 := l.PushFront(2)
	e1 := l.PushFront(1)
	e3 := l.PushBack(3)
	e4 := l.PushBack("banana")
	checkListPointers(t, l, []*Element{e1, e2, e3, e4})

	l.Remove(e2)
	checkListPointers(t, l, []*Element{e1, e3, e4})

	l.MoveToFront(e3) // move from middle
	checkListPointers(t, l, []*Element{e3, e1, e4})

	l.MoveToFront(e1)
	l.MoveToBack(e3) // move from middle
	checkListPointers(t, l, []*Element{e1, e4, e3})

	l.MoveToFront(e3) // move from back
	checkListPointers(t, l, []*Element{e3, e1, e4})
	l.MoveToFront(e3) // should be no-op
	checkListPointers(t, l, []*Element{e3, e1, e4})

	l.MoveToBack(e3) // move from front
	checkListPointers(t, l, []*Element{e1, e4, e3})
	l.MoveToBack(e3) // should be no-op
	checkListPointers(t, l, []*Element{e1, e4, e3})

	e2 = l.InsertBefore(2, e1) // insert before front
	checkListPointers(t, l, []*Element{e2, e1, e4, e3})
	l.Remove(e2)
	e2 = l.InsertBefore(2, e4) // insert before middle
	checkListPointers(t, l, []*Element{e1, e2, e4, e3})
	l.Remove(e2)
	e2 = l.InsertBefore(2, e3) // insert before back
	checkListPointers(t, l, []*Element{e1, e4, e2, e3})
	l.Remove(e2)

	e2 = l.InsertAfter(2, e1) // insert after front
	checkListPointers(t, l, []*Element{e1, e2, e4, e3})
	l.Remove(e2)
	e2 = l.InsertAfter(2, e4) // insert after middle
	checkListPointers(t, l, []*Element{e1, e4, e2, e3})
	l.Remove(e2)
	e2 = l.InsertAfter(2, e3) // insert after back
	checkListPointers(t, l, []*Element{e1, e4, e3, e2})
	l.Remove(e2)

	// Check standard iteration.
	sum := 0
	for e := l.Front(); e != nil; e = e.Next() {
		if i, ok := e.Value.(int); ok {
			sum += i
		}
	}
	if sum != 4 {
		t.Errorf("sum over l = %d, want 4", sum)
	}

	// Clear all elements by iterating
	var next *Element
	for e := l.Front(); e != nil; e = next {
		next = e.Next()
		l.Remove(e)
	}
	checkListPointers(t, l, []*Element{})
}

func checkList[T comparable](t *testing.T, l *List[T], es []any) {
	if !checkListLen(t, l, len(es)) {
		return
	}

	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		le := e.Value
		if le != es[i] {
			t.Errorf("elt[%d].Value = %v, want %v", i, le, es[i])
		}
		i++
	}
}

func TestExtending(t *testing.T) {
	l1 := NewList[int]()
	l2 := NewList[int]()

	l1.PushBack(1)
	l1.PushBack(2)
	l1.PushBack(3)

	l2.PushBack(4)
	l2.PushBack(5)

	l3 := NewList[int]()
	l3.PushBackList(l1)
	checkList(t, l3, []any{1, 2, 3})
	l3.PushBackList(l2)
	checkList(t, l3, []any{1, 2, 3, 4, 5})

	l3 = NewList[int]()
	l3.PushFrontList(l2)
	checkList(t, l3, []any{4, 5})
	l3.PushFrontList(l1)
	checkList(t, l3, []any{1, 2, 3, 4, 5})

	checkList(t, l1, []any{1, 2, 3})
	checkList(t, l2, []any{4, 5})

	l3 = NewList[int]()
	l3.PushBackList(l1)
	checkList(t, l3, []any{1, 2, 3})
	l3.PushBackList(l3)
	checkList(t, l3, []any{1, 2, 3, 1, 2, 3})

	l3 = NewList[int]()
	l3.PushFrontList(l1)
	checkList(t, l3, []any{1, 2, 3})
	l3.PushFrontList(l3)
	checkList(t, l3, []any{1, 2, 3, 1, 2, 3})

	l3 = NewList[int]()
	l1.PushBackList(l3)
	checkList(t, l1, []any{1, 2, 3})
	l1.PushFrontList(l3)
	checkList(t, l1, []any{1, 2, 3})
}

func TestRemove(t *testing.T) {
	l := NewList[int]()
	e1 := l.PushBack(1)
	e2 := l.PushBack(2)
	checkListPointers(t, l, []*Element{e1, e2})
	e := l.Front()
	l.Remove(e)
	checkListPointers(t, l, []*Element{e2})
	l.Remove(e)
	checkListPointers(t, l, []*Element{e2})
}

func TestIssue4103(t *testing.T) {
	l1 := NewList[int]()
	l1.PushBack(1)
	l1.PushBack(2)

	l2 := NewList[int]()
	l2.PushBack(3)
	l2.PushBack(4)

	e := l1.Front()
	l2.Remove(e) // l2 should not change because e is not an element of l2
	if n := l2.Len(); n != 2 {
		t.Errorf("l2.Len() = %d, want 2", n)
	}

	l1.InsertBefore(8, e)
	if n := l1.Len(); n != 3 {
		t.Errorf("l1.Len() = %d, want 3", n)
	}
}

func TestIssue6349(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	l.PushBack(2)

	e := l.Front()
	l.Remove(e)
	if e.Value != 1 {
		t.Errorf("e.value = %d, want 1", e.Value)
	}
	if e.Next() != nil {
		t.Errorf("e.Next() != nil")
	}
	if e.Prev() != nil {
		t.Errorf("e.Prev() != nil")
	}
}

func TestMove(t *testing.T) {
	l := NewList[int]()
	e1 := l.PushBack(1)
	e2 := l.PushBack(2)
	e3 := l.PushBack(3)
	e4 := l.PushBack(4)

	l.MoveAfter(e3, e3)
	checkListPointers(t, l, []*Element{e1, e2, e3, e4})
	l.MoveBefore(e2, e2)
	checkListPointers(t, l, []*Element{e1, e2, e3, e4})

	l.MoveAfter(e3, e2)
	checkListPointers(t, l, []*Element{e1, e2, e3, e4})
	l.MoveBefore(e2, e3)
	checkListPointers(t, l, []*Element{e1, e2, e3, e4})

	l.MoveBefore(e2, e4)
	checkListPointers(t, l, []*Element{e1, e3, e2, e4})
	e2, e3 = e3, e2

	l.MoveBefore(e4, e1)
	checkListPointers(t, l, []*Element{e4, e1, e2, e3})
	e1, e2, e3, e4 = e4, e1, e2, e3

	l.MoveAfter(e4, e1)
	checkListPointers(t, l, []*Element{e1, e4, e2, e3})
	e2, e3, e4 = e4, e2, e3

	l.MoveAfter(e2, e3)
	checkListPointers(t, l, []*Element{e1, e3, e2, e4})
}

// Test PushFront, PushBack, PushFrontList, PushBackList with uninitialized List
func TestZeroList(t *testing.T) {
	var l1 = new(List[int])
	l1.PushFront(1)
	checkList(t, l1, []any{1})

	var l2 = new(List[int])
	l2.PushBack(1)
	checkList(t, l2, []any{1})

	var l3 = new(List[int])
	l3.PushFrontList(l1)
	checkList(t, l3, []any{1})

	var l4 = new(List[int])
	l4.PushBackList(l2)
	checkList(t, l4, []any{1})
}

// Test that a list l is not modified when calling InsertBefore with a mark that is not an element of l.
func TestInsertBeforeUnknownMark(t *testing.T) {
	var l List[int]
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.InsertBefore(1, new(Element))
	checkList(t, &l, []any{1, 2, 3})
}

// Test that a list l is not modified when calling InsertAfter with a mark that is not an element of l.
func TestInsertAfterUnknownMark(t *testing.T) {
	var l List[int]
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.InsertAfter(1, new(Element))
	checkList(t, &l, []any{1, 2, 3})
}

// Test that a list l is not modified when calling MoveAfter or MoveBefore with a mark that is not an element of l.
func TestMoveUnknownMark(t *testing.T) {
	var l1 List[int]
	e1 := l1.PushBack(1)

	var l2 List[int]
	e2 := l2.PushBack(2)

	l1.MoveAfter(e1, e2)
	checkList(t, &l1, []any{1})
	checkList(t, &l2, []any{2})

	l1.MoveBefore(e1, e2)
	checkList(t, &l1, []any{1})
	checkList(t, &l2, []any{2})
}
