package gopool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Ring struct {
	sem     int32
	size    int32
	workers *RingBuffer
	queued  chan func()
	cond    *sync.Cond

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

func NewRing(size, queue, spawn int) *Ring {
	if spawn <= 0 {
		panic("dead queue configuration detected")
	}
	if spawn > size {
		panic("spawn > workers")
	}
	if queue > 1 {
		queue -= 1
	}
	p := &Ring{
		size:    int32(size),
		workers: NewRingBuffer(uint64(size)),
		queued:  make(chan func(), queue),
		cond:    sync.NewCond(&sync.Mutex{}),
		stop:    make(chan struct{}),
	}
	for i := 0; i < spawn; i++ {
		atomic.AddInt32(&p.sem, 1)
		p.workers.Put(p.newWorker())
	}
	go p.runQueued()
	return p
}

func (p *Ring) newWorker() *worker {
	atomic.AddInt32(&p.active, 1)
	p.wait.Add(1)
	w := &worker{
		tasks: make(chan func(), 1),
	}
	go p.runWorker(w)
	return w
}

func (p *Ring) runWorker(w *worker) {
	defer p.doneWorker()
	threshold := p.size - int32(runtime.NumCPU()*2)
L1:
	for {
		select {
		case task := <-w.tasks:
			task()
			p.putWorker(w)
			busy := atomic.AddInt32(&p.busy, -1)
			if busy >= threshold {
				p.cond.Signal()
			}
		case <-p.stop:
			p.cond.Signal()
			break L1
		}
	}
L2:
	for {
		select {
		case task := <-w.tasks:
			task()
			atomic.AddInt32(&p.busy, -1)
		case task, ok := <-p.queued:
			if !ok {
				break L2
			}
			task()
			atomic.AddInt32(&p.pending, -1)
		}
	}
}

func (p *Ring) doneWorker() {
	atomic.AddInt32(&p.active, -1)
	atomic.AddInt32(&p.sem, -1)
	p.wait.Done()
}

func (p *Ring) putWorker(w *worker) {
	p.workers.Put(w)
}

func (p *Ring) Schedule(task func()) error {
	return p.schedule(task, 0)
}

func (p *Ring) ScheduleTimeout(task func(), timeout time.Duration) error {
	return p.schedule(task, timeout)
}

func (p *Ring) schedule(task func(), timeout time.Duration) (err error) {
	if atomic.LoadInt32(&p.isStopped) > 0 {
		return ErrStopped
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
		if err == ErrTimeout {
			atomic.AddUint64(&p.timeout, 1)
		}
		return err
	}
	w.(*worker).submit(task)
	return nil
}

func (p *Ring) scheduleQueue(task func(), timeout time.Duration) error {
	var deadline <-chan time.Time
	if timeout > 0 {
		deadline = time.After(timeout)
	}

	atomic.AddInt32(&p.pending, 1)
	select {
	case <-deadline:
		atomic.AddInt32(&p.pending, -1)
		atomic.AddUint64(&p.timeout, 1)
		return ErrTimeout
	case <-p.stop:
		atomic.AddInt32(&p.pending, -1)
		return ErrStopped
	case p.queued <- task:
		return nil
	}
}

func (p *Ring) runQueued() {
	var task func()
L1:
	for atomic.LoadInt32(&p.isStopped) == 0 {
		select {
		case task = <-p.queued:
			p.doQueuedTask(task)
		case <-p.stop:
			break L1
		}
	}
L2:
	// done the buffered tasks
	for {
		select {
		case task = <-p.queued:
			task()
			atomic.AddInt32(&p.pending, -1)
		default:
			close(p.queued)
			break L2
		}
	}
}

func (p *Ring) doQueuedTask(task func()) {
	// Create new worker if available when all workers are busy.
	if atomic.LoadInt32(&p.sem) < p.size {
		if atomic.LoadInt32(&p.busy) >= atomic.LoadInt32(&p.active) {
			atomic.AddInt32(&p.sem, 1)
			atomic.AddInt32(&p.busy, 1)
			w := p.newWorker()
			w.submit(task)
			atomic.AddInt32(&p.pending, -1)
			return
		}
	}

	// Wait for a free worker to submit the task.
	for {
		p.cond.L.Lock()
		for atomic.LoadInt32(&p.busy) >= p.size {
			p.cond.Wait()
		}
		busy := atomic.AddInt32(&p.busy, 1)
		p.cond.L.Unlock()
		if busy > p.size {
			atomic.AddInt32(&p.busy, -1)
			runtime.Gosched()
			continue
		}
		p.scheduleRing(task, 0)
		atomic.AddInt32(&p.pending, -1)
		break
	}
}

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
