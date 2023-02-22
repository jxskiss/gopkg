package mselect

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
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
	// Submit submits a Task to the task executor.
	// After a Task's channel being closed, the task will be
	// automatically removed.
	// Calling this is a no-op after Stop is called.
	Submit(task *Task)

	// Count returns the count of running select tasks.
	// It always returns 0 after Stop is called.
	Count() int

	// Stop stops the task executor.
	Stop()
}

// New creates a new ManySelect.
func New() ManySelect {
	msel := &manySelect{
		tasks: make(chan any, 1),
	}
	return msel
}

type manySelect struct {
	mu      sync.Mutex
	buckets []*taskBucket

	tasks chan any // *Task

	count   int32
	stopped int32
}

func (p *manySelect) Submit(task *Task) {
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
	if atomic.LoadInt32(&p.stopped) > 0 {
		return 0
	}
	ret := int(atomic.LoadInt32(&p.count))
	if ret < 0 {
		ret = 0
	}
	return ret
}

func (p *manySelect) Stop() {
	atomic.StoreInt32(&p.stopped, 1)
	close(p.tasks)
}

func (p *manySelect) cap() int32 {
	bNum := len(p.buckets)
	return int32(bNum*bucketSize - bNum)
}

func (p *manySelect) decrCount() {
	atomic.AddInt32(&p.count, -1)
}

//go:noescape
//go:linkname reflect_rselect reflect.rselect
//nolint:all
func reflect_rselect([]runtimeSelect) (chosen int, recvOK bool)

// A runtimeSelect is a single case passed to reflect_rselect.
// This must match reflect.runtimeSelect.
type runtimeSelect struct {
	Dir reflect.SelectDir // SelectSend, SelectRecv or SelectDefault
	Typ unsafe.Pointer    // *rtype, channel type
	Ch  unsafe.Pointer    // channel
	Val unsafe.Pointer    // ptr to data (SendDir) or ptr to receive buffer (RecvDir)
}
