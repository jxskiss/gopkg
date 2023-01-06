package mselect

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ManySelect interface {
	Submit(task *Task)
	Count() int
	Stop()
}

func New() ManySelect {
	msel := &manySelect{
		tasks: make(chan interface{}, 1),
	}
	return msel
}

type manySelect struct {
	mu      sync.Mutex
	buckets []*taskBucket

	tasks chan interface{} // *Task

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
	return int(atomic.LoadInt32(&p.count))
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
func reflect_rselect([]runtimeSelect) (chosen int, recvOK bool)

// A runtimeSelect is a single case passed to reflect_rselect.
// This must match reflect.runtimeSelect.
type runtimeSelect struct {
	Dir reflect.SelectDir // SelectSend, SelectRecv or SelectDefault
	Typ unsafe.Pointer    // *rtype, channel type
	Ch  unsafe.Pointer    // channel
	Val unsafe.Pointer    // ptr to data (SendDir) or ptr to receive buffer (RecvDir)
}
