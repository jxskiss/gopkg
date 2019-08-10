// Package gopool contains tools for goroutine reuse.
package gopool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrTimeout returned by Pool to indicate that there are no free
	// goroutines during some period of time.
	ErrTimeout = errors.New("gopool: schedule timeout")
	ErrStopped = errors.New("gopool: pool stopped")
)

// Pool contains logic of goroutine reuse.
type Pool struct {
	sem  chan struct{}
	work chan func()

	wait      sync.WaitGroup
	stop      chan struct{}
	isStopped int32

	// stats
	active  int32
	busy    int32
	pending int32
	timeout uint64
}

// NewPool creates new goroutine pool with given size. It also creates a work
// queue of given size. Finally, it spawns given amount of goroutines
// immediately.
func NewPool(size, queue, spawn int) *Pool {
	if spawn <= 0 && queue > 0 {
		panic("dead queue configuration detected")
	}
	if spawn > size {
		panic("spawn > workers")
	}
	p := &Pool{
		sem:  make(chan struct{}, size),
		work: make(chan func(), queue),
		stop: make(chan struct{}),
	}
	for i := 0; i < spawn; i++ {
		p.sem <- struct{}{}
		p.wait.Add(1)
		go p.worker(func() {})
	}

	return p
}

// Schedule schedules task to be executed over pool's workers.
func (p *Pool) Schedule(task func()) error {
	return p.schedule(task, nil)
}

// ScheduleTimeout schedules task to be executed over pool's workers.
// It returns ErrTimeout when no free workers met during given timeout.
func (p *Pool) ScheduleTimeout(task func(), timeout time.Duration) error {
	return p.schedule(task, time.After(timeout))
}

func (p *Pool) schedule(task func(), timeout <-chan time.Time) (err error) {
	if atomic.LoadInt32(&p.isStopped) == 1 {
		return ErrStopped
	}
	select {
	case <-p.stop:
		return ErrStopped
	case <-timeout:
		atomic.AddUint64(&p.timeout, 1)
		return ErrTimeout
	case p.work <- task:
		atomic.AddInt32(&p.pending, 1)
	case p.sem <- struct{}{}:
		p.wait.Add(1)
		go p.worker(nil)
	}
	return nil
}

func (p *Pool) worker(task func()) {
	atomic.AddInt32(&p.active, 1)
	defer p.doneWorker()

	if task != nil {
		atomic.AddInt32(&p.busy, 1)
		task()
		atomic.AddInt32(&p.busy, -1)
	}

	for {
		select {
		case <-p.stop:
			return
		case task := <-p.work:
			atomic.AddInt32(&p.pending, -1)
			atomic.AddInt32(&p.busy, 1)
			task()
			atomic.AddInt32(&p.busy, -1)
		}
	}
}

func (p *Pool) doneWorker() {
	atomic.AddInt32(&p.active, -1)
	p.wait.Done()
	<-p.sem
}

func (p *Pool) Stop() {
	if !atomic.CompareAndSwapInt32(&p.isStopped, 0, 1) {
		return
	}

	// clean up the pending tasks
	p.wait.Add(1)
	go func() {
		defer p.wait.Done()
		for {
			select {
			case task := <-p.work:
				atomic.AddInt32(&p.pending, -1)
				atomic.AddInt32(&p.busy, 1)
				task()
				atomic.AddInt32(&p.busy, -1)
			default:
				close(p.stop)
				return
			}
		}
	}()
	p.wait.Wait()
}

func (p *Pool) Stats() (active, busy, pending, timeout int) {
	active = int(atomic.LoadInt32(&p.active))
	busy = int(atomic.LoadInt32(&p.busy))
	pending = int(atomic.LoadInt32(&p.pending))
	timeout = int(atomic.SwapUint64(&p.timeout, 0))
	return
}
