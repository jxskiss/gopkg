package mselect

import (
	"reflect"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

// NewTask creates a new Task which can be submitted to ManySelect.
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
		convFunc: func(p unsafe.Pointer) any {
			return *(*T)(p)
		},
	}
	return task
}

// Task is a channel receiving task which can be submitted to ManySelect.
// A zero Task is not ready to use, use NewTask to create a Task.
type Task struct {
	ch       reflect.Value
	execFunc func(v unsafe.Pointer, ok bool)
	newFunc  func() unsafe.Pointer
	convFunc func(p unsafe.Pointer) any
}

func (t *Task) newRuntimeSelect() runtimeSelect {
	rtype := reflectx.ToRType(t.ch.Type())
	rsel := runtimeSelect{
		Dir: reflect.SelectRecv,
		Typ: unsafe.Pointer(rtype),
		Ch:  t.ch.UnsafePointer(),
		Val: t.newFunc(),
	}
	return rsel
}

func (t *Task) getAndResetRecvValue(rsel *runtimeSelect) unsafe.Pointer {
	recv := rsel.Val
	rsel.Val = t.newFunc()
	return recv
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
