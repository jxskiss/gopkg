package gopool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type unboundedPool struct {
	maxAge  time.Duration
	maxIdle int32

	workers   atomic.Int32
	idleCount atomic.Int32
	unixMilli atomic.Int64

	taskCh     chan *task
	tickerOnce sync.Once
	killSig    atomic.Value

	panicHandler func(ctx context.Context, r any)
	runner       taskRunner
}

func newUnboundedPool(config *Config, runner taskRunner) *unboundedPool {
	config.checkAndSetUnboundedDefaults()
	up := &unboundedPool{
		taskCh:       make(chan *task),
		maxAge:       config.maxAge,
		maxIdle:      config.maxIdle,
		panicHandler: config.PanicHandler,
		runner:       runner,
	}
	return up
}

func (p *unboundedPool) submitTask(ctx context.Context, arg any) {
	// Inline newTask(ctx, arg)
	var t *task
	if tt := taskPool.Get(); tt != nil {
		t = tt.(*task)
	} else {
		t = &task{}
	}
	*t = task{ctx: ctx, arg: arg}

	select {
	case p.taskCh <- t: // got a free worker
		return
	default:
		// all workers are busy, start a new worker
		go p.runWorker(t)
	}
}

func (p *unboundedPool) workersCount() int {
	return int(p.workers.Load())
}

func (p *unboundedPool) runWorker(task *task) {
	p.tickerOnce.Do(func() {
		p.killSig.Store(make(chan struct{}))
		go p.runTicker()
	})

	p.workers.Add(1)
	defer p.workers.Add(-1)

	p.runner(task, p.panicHandler)

	// check for maxIdle constraint
	if p.idleCount.Load() >= p.maxIdle {
		return
	}

	// idle, wait for new tasks
	startAt := time.Now().UnixMilli()
	killSig := p.killSig.Load().(chan struct{})
	p.idleCount.Add(1)
	for {
		select {
		case t := <-p.taskCh:
			p.idleCount.Add(-1)
			p.runner(t, p.panicHandler)
			p.idleCount.Add(1)
		case <-killSig:
			now := p.unixMilli.Load()
			if now-startAt > int64(p.maxAge/time.Millisecond) {
				p.idleCount.Add(-1)
				return
			}
			killSig = p.killSig.Load().(chan struct{})
		}
	}
}

func (p *unboundedPool) runTicker() {
	// Split maxAge to 1/1000, but at least 20ms.
	interval := max(p.maxAge/1000, 20*time.Millisecond)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	milli := time.Now().UnixMilli()
	p.unixMilli.Store(milli)

	pre := milli
	for now := range ticker.C {
		milli = now.UnixMilli()
		p.unixMilli.Store(milli)
		if milli-pre > 1000 { // per second
			pre = milli
			old := p.killSig.Swap(make(chan struct{}))
			close(old.(chan struct{}))
		}
	}
}
