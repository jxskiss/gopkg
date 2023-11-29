package mselect

import (
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

const (
	sigTaskNum = 2
	bucketSize = 256
	bucketCap  = bucketSize - sigTaskNum
)

var (
	blockCh   chan any
	blockTask = NewTask(blockCh, nil, nil)
)

type taskBucket struct {
	m *manySelect

	cases []linkname.RuntimeSelect
	tasks []*Task

	block bool
}

func newTaskBucket(msel *manySelect, userTask *Task) *taskBucket {
	b := &taskBucket{
		m:     msel,
		cases: make([]linkname.RuntimeSelect, 0, 16),
		tasks: make([]*Task, 0, 16),
	}
	sigTask := msel.sigTask
	stopTask := NewTask(b.m.stop, nil, nil)
	b.addTask(sigTask)
	b.addTask(stopTask)
	b.addTask(userTask)
	go b.loop()
	return b
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
			b.processSignal(recv, ok)
			if !ok {
				return // stopped
			}
		case 1: // stopped
			b.stop()
			return
		default:
			b.processTask(i, task, recv, ok)
		}
	}
}

func (b *taskBucket) processSignal(recv unsafe.Pointer, ok bool) {
	if !ok { // stopped
		b.stop()
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

func (b *taskBucket) processTask(i int, task *Task, recv unsafe.Pointer, ok bool) {
	// Execute the task first.
	// When the channel is closed, call callback functions with
	// a zero value and ok = false.
	if task.execFunc != nil {
		task.execFunc(recv, ok)
	}

	// Delete task if the channel was closed.
	if !ok {
		b.deleteTask(i)
		b.m.decrCount(1)

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
	b.tasks = append(b.tasks, task)
	b.cases = append(b.cases, task.newRuntimeSelect())
}

func (b *taskBucket) deleteTask(i int) {
	n := len(b.cases)
	if n > sigTaskNum+1 {
		b.cases[i] = b.cases[n-1]
		b.tasks[i] = b.tasks[n-1]
	}
	b.cases[n-1] = linkname.RuntimeSelect{}
	b.tasks[n-1] = nil
	b.cases = b.cases[:n-1]
	b.tasks = b.tasks[:n-1]
}

func (b *taskBucket) stop() {
	n := len(b.tasks) - sigTaskNum // don't count the signal tasks
	b.m.decrCount(n)
	b.cases = nil
	b.tasks = nil

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
