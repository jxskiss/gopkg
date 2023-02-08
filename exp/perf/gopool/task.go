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

func (l *taskList) pop() (t *task) {
	l.mu.Lock()
	if l.head != nil {
		t = l.head
		l.head = l.head.next
		l.count++
	}
	l.mu.Unlock()
	return
}
