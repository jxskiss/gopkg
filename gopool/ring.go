package gopool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	queuepkg "github.com/Workiva/go-datastructures/queue"
)

type Ring struct {
	sem     int32
	size    int32
	workers *queuepkg.RingBuffer
	queued  chan func()
	sleepf  func()

	wait      sync.WaitGroup
	stop      chan struct{}
	isStopped int32

	// stats
	active  int32
	busy    int32
	pending int32
	timeout uint64
}

type worker struct {
	tasks chan func()
}

func (w *worker) submit(task func()) {
	w.tasks <- task
}

func NewRing(size, queue, spawn int, sleepFunc func()) *Ring {
	if spawn <= 0 {
		panic("dead queue configuration detected")
	}
	if spawn > size {
		panic("spawn > workers")
	}
	if queue > 1 {
		queue -= 1
	}
	if sleepFunc == nil {
		sleepFunc = func() { runtime.Gosched() }
	}
	p := &Ring{
		size:    int32(size),
		queued:  make(chan func(), queue),
		workers: queuepkg.NewRingBuffer(uint64(size)),
		stop:    make(chan struct{}),
		sleepf:  sleepFunc,
	}
	for i := 0; i < spawn; i++ {
		atomic.AddInt32(&p.sem, 1)
		p.wait.Add(1)
		p.workers.Put(p.newWorker())
	}
	go p.runQueued()
	return p
}

func (p *Ring) newWorker() *worker {
	atomic.AddInt32(&p.active, 1)
	w := &worker{
		tasks: make(chan func(), 1),
	}
	go p.runWorker(w)
	return w
}

func (p *Ring) runWorker(w *worker) {
	defer p.doneWorker()
LOOP:
	for {
		select {
		case task := <-w.tasks:
			task()
			err := p.putWorker(w)
			atomic.AddInt32(&p.busy, -1)
			if err != nil {
				break LOOP
			}
		case <-p.stop:
			// After pool stopped, there will be at most one task left.
			select {
			case task := <-w.tasks:
				task()
				atomic.AddInt32(&p.busy, -1)
			default:
			}
			break LOOP
		}
	}
}

func (p *Ring) doneWorker() {
	atomic.AddInt32(&p.active, -1)
	atomic.AddInt32(&p.sem, -1)
	p.wait.Done()
}

func (p *Ring) putWorker(w *worker) (err error) {
	err = p.workers.Put(w)
	return err
}

func (p *Ring) Schedule(task func()) error {
	return p.schedule(task, 0)
}

func (p *Ring) ScheduleTimeout(task func(), timeout time.Duration) error {
	return p.schedule(task, timeout)
}

func (p *Ring) schedule(task func(), timeout time.Duration) (err error) {
	if atomic.LoadInt32(&p.isStopped) > 0 {
		return ErrPoolStopped
	}

	// tasks are queued
	if atomic.LoadInt32(&p.pending) > 0 {
		err = p.scheduleQueue(task, timeout)
		return
	}

	busy := atomic.AddInt32(&p.busy, 1)
	if busy < p.size && busy < atomic.LoadInt32(&p.active) {
		// free worker available, send directly to the ring queue
		err = p.scheduleRing(task, timeout)
		if err != nil {
			atomic.AddInt32(&p.busy, -1)
		}
		return
	} else {
		// workers are busy, queue the task
		atomic.AddInt32(&p.busy, -1)
		err = p.scheduleQueue(task, timeout)
		return
	}
}

func (p *Ring) scheduleRing(task func(), timeout time.Duration) error {
	w, err := p.workers.Poll(timeout)
	if err != nil {
		if err == queuepkg.ErrTimeout {
			atomic.AddUint64(&p.timeout, 1)
			return ErrScheduleTimeout
		}
		if err == queuepkg.ErrDisposed {
			return ErrPoolStopped
		}
		return err
	}
	w.(*worker).submit(task)
	return nil
}

func (p *Ring) scheduleQueue(task func(), timeout time.Duration) error {
	// TODO: should we check pool stopped?
	if timeout == 0 {
		p.queued <- task
	} else {
		select {
		case <-time.After(timeout):
			atomic.AddUint64(&p.timeout, 1)
			return ErrScheduleTimeout
		case p.queued <- task:
		}
	}
	atomic.AddInt32(&p.pending, 1)
	return nil
}

func (p *Ring) runQueued() {
	// TODO: check pool stopped?
	for task := range p.queued {
		// Create new worker if available when all workers are busy.
		if atomic.LoadInt32(&p.sem) < p.size {
			if atomic.LoadInt32(&p.busy) >= atomic.LoadInt32(&p.active) {
				atomic.AddInt32(&p.sem, 1)
				atomic.AddInt32(&p.busy, 1)
				p.wait.Add(1)
				w := p.newWorker()
				w.submit(task)
				atomic.AddInt32(&p.pending, -1)
				continue
			}
		}

		for {
			if atomic.LoadInt32(&p.busy) >= p.size {
				p.sleepf()
				continue
			}
			busy := atomic.AddInt32(&p.busy, 1)
			if busy > p.size {
				atomic.AddInt32(&p.busy, -1)
				runtime.Gosched()
				continue
			}
			// got free worker, submit the task
			err := p.scheduleRing(task, 0)
			if err != nil {
				// poll has been stopped
				atomic.AddInt32(&p.busy, -1)
				return
			}
			atomic.AddInt32(&p.pending, -1)
			break
		}
	}
}

// TODO: wait queued tasks.
func (p *Ring) Stop() {
	if !atomic.CompareAndSwapInt32(&p.isStopped, 0, 1) {
		return
	}
	close(p.stop)
	p.wait.Wait()
}

func (p *Ring) Stats() (active, busy, pending, timeout int) {
	active = int(atomic.LoadInt32(&p.active))
	busy = int(atomic.LoadInt32(&p.busy))
	if busy > active {
		busy--
	}
	pending = int(atomic.LoadInt32(&p.pending))
	timeout = int(atomic.SwapUint64(&p.timeout, 0))
	return
}
