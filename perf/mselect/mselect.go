package mselect

import (
	"sync"
	"sync/atomic"
)

// ManySelect is a channel receiving task executor,
// it runs many channel receiving operations simultaneously,
// tasks can be added dynamically.
//
// It is not designed for traffic-heavy channels,
// instead when there are many long-living goroutines waiting on
// channels with very little traffic,
// they do some simple things when a value is received from a channel,
// you may use this to avoid running a lot of goroutines.
type ManySelect interface {
	// Add submits a Task to the task executor.
	// After a Task's channel being closed, the task will be automatically removed.
	// Calling this is a no-op after Stop is called.
	Add(task *Task)

	// Count returns the count of running select tasks.
	Count() int

	// Stop stops the task executor.
	Stop()
}

// New creates a new ManySelect.
func New() ManySelect {
	msel := &manySelect{
		tasks: make(chan *Task, 1),
		stop:  make(chan struct{}),
	}
	msel.sigTask = NewTask(msel.tasks, nil, nil)
	return msel
}

type manySelect struct {
	mu      sync.Mutex
	buckets []*taskBucket

	tasks   chan *Task
	sigTask *Task

	stop    chan struct{}
	stopped int32

	count int32
}

func (p *manySelect) Add(task *Task) {
	if atomic.LoadInt32(&p.stopped) > 0 {
		return
	}

	p.mu.Lock()
	if atomic.AddInt32(&p.count, 1) < p.cap() {
		p.mu.Unlock()
		p.tasks <- task
		return
	}

	nb := newTaskBucket(p, task)
	p.buckets = append(p.buckets, nb)
	p.mu.Unlock()
}

func (p *manySelect) Count() int {
	ret := int(atomic.LoadInt32(&p.count))
	if ret < 0 {
		ret = 0
	}
	return ret
}

func (p *manySelect) Stop() {
	atomic.StoreInt32(&p.stopped, 1)
	close(p.stop)

	// Note that we don't close p.tasks here since we may have
	// concurrent senders running p.Add.
}

func (p *manySelect) cap() int32 {
	bNum := len(p.buckets)
	return int32(bNum * bucketCap)
}

func (p *manySelect) decrCount(n int) {
	atomic.AddInt32(&p.count, -int32(n))
}
