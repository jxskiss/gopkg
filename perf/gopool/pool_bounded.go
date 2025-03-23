package gopool

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
)

type boundedPool struct {
	scaleThreshold int

	// taskCh sends tasks to permanent workers.
	taskCh chan *task

	// mu protects adhocState and taskList.
	// adhocState:
	// - higher 32 bits is adhocLimit, max number of adhoc workers that can run simultaneously
	// - lower 32 bits is adhocCount, the number of currently running adhoc workers
	mu         sync.Mutex
	adhocState int64
	taskList   taskList

	panicHandler func(ctx context.Context, r any)
	runner       taskRunner
}

func newBoundedPool(config *Config, runner taskRunner) *boundedPool {
	config.checkAndSetBoundedDefaults()
	bp := &boundedPool{
		scaleThreshold: config.scaleThreshold,
		panicHandler:   config.PanicHandler,
		runner:         runner,
	}
	if config.permanentWorkerNum > 0 {
		bp.startPermanentWorkers(config.permanentWorkerNum)
	}
	bp.setAdhocWorkerLimit(config.adhocWorkerLimit)
	return bp
}

func (p *boundedPool) setAdhocWorkerLimit(limit int) {
	if limit <= 0 || limit > math.MaxInt32 {
		limit = math.MaxInt32
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	oldLimit, _ := p.getAdhocState()
	diff := int32(limit) - oldLimit
	if diff != 0 {
		atomic.AddInt64(&p.adhocState, int64(diff)<<32)
	}
}

func (p *boundedPool) getAdhocState() (limit, count int32) {
	x := atomic.LoadInt64(&p.adhocState)
	return int32(x >> 32), int32((x << 32) >> 32)
}

func (p *boundedPool) incAdhocWorkerCount() {
	atomic.AddInt64(&p.adhocState, 1)
}

func (p *boundedPool) decAdhocWorkerCount() {
	atomic.AddInt64(&p.adhocState, -1)
}

func (p *boundedPool) submitTask(ctx context.Context, arg any) {
	// Inline newTask(ctx, arg)
	var t *task
	if tt := taskPool.Get(); tt != nil {
		t = tt.(*task)
	} else {
		t = &task{}
	}
	*t = task{ctx: ctx, arg: arg}

	// Try permanent worker first.
	if p.taskCh != nil {
		select {
		case p.taskCh <- t:
			return
		default:
		}
	}

	// No free permanent worker available, check to start a new
	// adhoc worker or submit to the task list.
	//
	// Start a new adhoc worker if there are currently no adhoc workers,
	// or the following two conditions are met:
	//   1. the number of tasks is greater than the threshold.
	//   2. The current number of adhoc workers is less than the limit.
	p.mu.Lock()
	tCnt := p.taskList.count + 1
	wLimit, wCnt := p.getAdhocState()
	if wCnt == 0 || (tCnt >= p.scaleThreshold && wCnt < wLimit) {
		p.incAdhocWorkerCount()
		p.mu.Unlock()
		go p.adhocWorker(t)
	} else {
		p.taskList.add(t)
		p.mu.Unlock()
	}
}

func (p *boundedPool) startPermanentWorkers(n int) {
	p.taskCh = make(chan *task)
	for i := 0; i < n; i++ {
		go p.permanentWorker()
	}
}

func (p *boundedPool) permanentWorker() {
	for {
		select {
		case t := <-p.taskCh:
			p.runner(t, p.panicHandler)

			// Drain pending tasks.
			for {
				p.mu.Lock()
				t = p.taskList.pop()
				p.mu.Unlock()
				if t == nil {
					break
				}
				p.runner(t, p.panicHandler)
			}
		}
	}
}

func (p *boundedPool) adhocWorker(t *task) {
	p.runner(t, p.panicHandler)
	for {
		p.mu.Lock()
		t = p.taskList.pop()
		if t == nil {
			p.decAdhocWorkerCount()
			p.mu.Unlock()
			return
		}
		p.mu.Unlock()
		p.runner(t, p.panicHandler)
	}
}
