package mselect

import (
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

const (
	sigTaskNum = 3
	bucketSize = 256
	bucketCap  = bucketSize - sigTaskNum
)

var (
	blockCh   chan any
	blockTask = NewTask(blockCh, nil, nil)
)

type taskBucket struct {
	m *manySelect

	delCh chan *Task

	cases []linkname.RuntimeSelect
	tasks []*Task

	block   bool
	stopped bool
}

func newTaskBucket(msel *manySelect, userTask *Task) *taskBucket {
	b := &taskBucket{
		m:     msel,
		delCh: make(chan *Task),
		cases: make([]linkname.RuntimeSelect, 0, 16),
		tasks: make([]*Task, 0, 16),
	}
	sigTask := msel.sigTask
	stopTask := NewTask(b.m.stop, nil, nil)
	delTask := NewTask(b.delCh, nil, nil)
	b.addTask(sigTask)
	b.addTask(stopTask)
	b.addTask(delTask)
	b.addTask(userTask)
	go b.loop()
	return b
}

func (b *taskBucket) signalDelete(task *Task) {
	b.delCh <- task
}

func (b *taskBucket) loop() {
	for {
		// Wait on select cases.
		i, ok := linkname.Reflect_rselect(b.cases)
		task := b.tasks[i]
		recv := task.getAndResetRecvValue(&b.cases[i])

		// Got a signal or a new task submitted.
		switch i {
		case 0: // signal
			if !ok { // stopped
				b.stop()
				return
			}
			b.processSignal(recv)
		case 1: // stopped
			b.stop()
			return
		case 2: // delete task
			delTask := *(**Task)(recv)
			b.deleteTask(delTask)
		default:
			b.executeTask(task, recv, ok)
		}
	}
}

func (b *taskBucket) processSignal(recv unsafe.Pointer) {
	if b.stopped {
		return
	}

	// Add a new task.
	newTask := *(**Task)(recv)
	b.addTask(newTask)

	// If the bucket is full, block the signal channel to avoid
	// accepting new tasks.
	if len(b.cases) == bucketSize {
		b.tasks[0] = blockTask
		b.cases[0] = blockTask.newRuntimeSelect()
		b.block = true
	}
}

func (b *taskBucket) executeTask(task *Task, recv unsafe.Pointer, ok bool) {
	// Execute the task first.
	// When the channel is closed, call callback functions with
	// a zero value and ok = false.
	if task.execFunc != nil {
		task.execFunc(recv, ok)
	}

	// Delete task if the channel was closed.
	if !ok {
		b.deleteTask(task)

		// Reset signal task to accept new tasks.
		if b.block && len(b.cases) < bucketSize {
			sigTask := b.m.sigTask
			b.tasks[0] = sigTask
			b.cases[0] = sigTask.newRuntimeSelect()
			b.block = false
		}
		return
	}
}

func (b *taskBucket) addTask(task *Task) {
	task.mu.Lock()
	defer task.mu.Unlock()

	if task.deleted {
		return
	}
	if task.added {
		panic("mselect: task already added")
	}

	// The fields added, bucket, tIdx are used to coordinate with deletion,
	// don't write these fields for signal tasks.
	if idx := len(b.tasks); idx >= sigTaskNum {
		task.added = true
		task.bucket = b
		task.tIdx = idx
	}

	b.tasks = append(b.tasks, task)
	b.cases = append(b.cases, task.newRuntimeSelect())
}

func (b *taskBucket) deleteTask(task *Task) {
	task.mu.Lock()
	defer task.mu.Unlock()

	if !task.added {
		return
	}
	if task.bucket == nil || task.tIdx < sigTaskNum {
		panic("mselect: invalid task to delete")
	}

	task.deleted = true
	b.removeTask(task.tIdx)
	b.m.decrCount(1)
}

func (b *taskBucket) removeTask(i int) {
	n := len(b.cases)
	if n > sigTaskNum+1 && i < n-1 {
		b.cases[i] = b.cases[n-1]
		b.tasks[i] = b.tasks[n-1]
		b.tasks[i].tIdx = i
	}
	b.cases[n-1] = linkname.RuntimeSelect{}
	b.tasks[n-1] = nil
	b.cases = b.cases[:n-1]
	b.tasks = b.tasks[:n-1]
}

func (b *taskBucket) stop() {
	if b.stopped {
		return
	}

	n := len(b.tasks) - sigTaskNum // don't count the signal tasks
	b.m.decrCount(n)
	b.cases = nil
	b.tasks = nil
	b.stopped = true

	// Drain tasks to un-block any goroutines blocked by sending tasks
	// to b.m.tasks.
	for {
		select {
		case <-b.m.tasks:
			continue
		default:
			return
		}
	}
}
