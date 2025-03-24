// Copyright 2021 ByteDance Inc.
// Copyright 2023 Shawn Wang <jxskiss@126.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gopool

import (
	"context"
	"sync"
)

var taskPool sync.Pool

type task struct {
	ctx context.Context
	arg any

	next *task
}

//nolint:unused
func newTask(ctx context.Context, arg any) *task {
	if t := taskPool.Get(); t != nil {
		t1 := t.(*task)
		*t1 = task{ctx: ctx, arg: arg}
		return t1
	}
	return &task{ctx: ctx, arg: arg}
}

func (t *task) Recycle() {
	*t = task{}
	taskPool.Put(t)
}

type taskList struct {
	count int
	head  *task
	tail  *task
}

// add adds a task to the tail of the taskList.
func (l *taskList) add(t *task) {
	if l.head == nil {
		l.head = t
		l.tail = t
	} else {
		l.tail.next = t
		l.tail = t
	}
	l.count++
}

// pop returns a task from the head of the taskList.
func (l *taskList) pop() (t *task) {
	if l.head != nil {
		t = l.head
		l.head = l.head.next
		l.count--
	}
	return t
}

type taskRunner func(t *task, panicHandler func(ctx context.Context, r any))

func funcTaskRunner(t *task, panicHandler func(ctx context.Context, r any)) {
	defer func() {
		if r := recover(); r != nil {
			panicHandler(t.ctx, r)
		}
	}()
	t.arg.(func())()
	t.Recycle()
}

func newTypedTaskRunner[T any](handler func(context.Context, T)) taskRunner {
	return func(t *task, panicHandler func(ctx context.Context, r any)) {
		defer func() {
			if r := recover(); r != nil {
				panicHandler(t.ctx, r)
			}
		}()
		handler(t.ctx, t.arg.(T))
		t.Recycle()
	}
}
