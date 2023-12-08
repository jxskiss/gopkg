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

type taskRunner func(p *internalPool, t *task)

func (p *internalPool) runPermanentWorker() {
	var lock *sync.Mutex
	for {
		select {
		case t := <-p.taskCh:
			p.runner(p, t)

			// Drain pending tasks.
			for {
				t, lock = p.taskList.pop()
				lock.Unlock()
				if t == nil {
					break
				}
				p.runner(p, t)
			}
		}
	}
}

func (p *internalPool) runAdhocWorker() {
	p.incWorkerCount()
	go func() {
		for {
			t, lock := p.taskList.pop()
			if t == nil {
				p.decWorkerCount()
				lock.Unlock()
				return
			}
			lock.Unlock()
			p.runner(p, t)
		}
	}()
}

func funcTaskRunner(p *internalPool, t *task) {
	defer func() {
		if r := recover(); r != nil {
			p.config.PanicHandler(t.ctx, r)
		}
	}()
	t.arg.(func())()
	t.Recycle()
}

func newTypedTaskRunner[T any](handler func(context.Context, T)) taskRunner {
	return func(p *internalPool, t *task) {
		defer func() {
			if r := recover(); r != nil {
				p.config.PanicHandler(t.ctx, r)
			}
		}()
		handler(t.ctx, t.arg.(T))
		t.Recycle()
	}
}

var taskPool sync.Pool

type task struct {
	ctx context.Context
	arg any

	next *task
}

func newTask() *task {
	if t := taskPool.Get(); t != nil {
		return t.(*task)
	}
	return &task{}
}

func (t *task) Recycle() {
	*t = task{}
	taskPool.Put(t)
}

type taskList struct {
	mu    sync.Mutex
	count int32
	head  *task
	tail  *task
}

func (l *taskList) add(t *task) (count int) {
	l.mu.Lock()
	if l.head == nil {
		l.head = t
		l.tail = t
	} else {
		l.tail.next = t
		l.tail = t
	}
	l.count++
	count = int(l.count)
	l.mu.Unlock()
	return
}

// pop acquired the lock and returns a task from the head of the taskList.
//
// Note that the caller takes responsibility to release the lock.
func (l *taskList) pop() (t *task, lock *sync.Mutex) {
	l.mu.Lock()
	if l.head != nil {
		t = l.head
		l.head = l.head.next
		l.count--
	}
	return t, &l.mu
}
