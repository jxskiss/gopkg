package mselect

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

// NewTask creates a new Task which can be submitted to ManySelect.
//
// If syncCallback or asyncCallback is not nil, or both not nil,
// when a value is received from ch, syncCallback is called synchronously,
// asyncCallback will be run asynchronously in a new goroutine.
// When ch is closed, non-nil syncCallback and asyncCallback will be called
// with a zero value of T and ok is false.
func NewTask[T any](
	ch <-chan T,
	syncCallback func(v T, ok bool),
	asyncCallback func(v T, ok bool),
) *Task {
	task := &Task{
		ch:       reflect.ValueOf(ch),
		execFunc: buildTaskFunc(syncCallback, asyncCallback),
		newFunc: func() unsafe.Pointer {
			return unsafe.Pointer(new(T))
		},
		tIdx: -1,
	}
	return task
}

func buildTaskFunc[T any](
	syncCallback func(v T, ok bool),
	asyncCallback func(v T, ok bool),
) func(v unsafe.Pointer, ok bool) {
	if syncCallback == nil && asyncCallback == nil {
		return nil
	}
	return func(v unsafe.Pointer, ok bool) {
		tVal := *(*T)(v)
		if syncCallback != nil {
			syncCallback(tVal, ok)
		}
		if asyncCallback != nil {
			go asyncCallback(tVal, ok)
		}
	}
}

// Task is a channel receiving task which can be submitted to ManySelect.
// A zero Task is not ready to use, use NewTask to create a Task.
//
// Task holds internal state data and shall not be reused,
// user should always use NewTask to create new task objects.
type Task struct {
	ch       reflect.Value
	execFunc func(v unsafe.Pointer, ok bool)
	newFunc  func() unsafe.Pointer

	mu      sync.Mutex
	bucket  *taskBucket
	tIdx    int // task index in bucket
	added   bool
	deleted bool
}

func (t *Task) newRuntimeSelect() linkname.RuntimeSelect {
	rtype := reflectx.ToRType(t.ch.Type())
	rsel := linkname.RuntimeSelect{
		Dir: reflect.SelectRecv,
		Typ: unsafe.Pointer(rtype),
		Ch:  t.ch.UnsafePointer(),
		Val: t.newFunc(),
	}
	return rsel
}

func (t *Task) getAndResetRecvValue(rsel *linkname.RuntimeSelect) unsafe.Pointer {
	recv := rsel.Val
	rsel.Val = t.newFunc()
	return recv
}

func (t *Task) signalDelete() {
	t.mu.Lock()
	if t.deleted {
		t.mu.Unlock()
		return
	}
	t.deleted = true
	bucket := t.bucket
	t.mu.Unlock()

	// No need to hold lock to send the signal.
	if bucket != nil {
		bucket.signalDelete(t)
	}
}
