// Copyright 2025 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gopool

import (
	"context"
	"sync/atomic"
	"time"
)

// GoPool is a goroutine pool that helps to reuse goroutines
// for better performance.
type GoPool struct {
	name string

	workers int32
	maxIdle int32
	maxAge  time.Duration

	panicHandler PanicHandler

	tasks     chan task
	unixMilli int64
}

type task struct {
	ctx context.Context
	f   func()
}

var noopTask = task{f: func() {}}

// New creates a new GoPool with the given name and options.
func New(name string, option *Option) *GoPool {
	if option == nil {
		option = DefaultOption()
	}
	p := &GoPool{
		name:         name,
		maxIdle:      int32(option.MaxIdleWorkers),
		maxAge:       option.WorkerMaxAge,
		panicHandler: option.PanicHandler,
		tasks:        make(chan task, option.TaskChanBuffer),
	}
	if p.panicHandler == nil {
		p.panicHandler = defaultPanicHandler
	}
	return p
}

// Go runs the given func in background.
func (p *GoPool) Go(f func()) {
	p.CtxGo(context.Background(), f)
}

// CtxGo runs the given func in background, and it passes ctx to panic handler when happens.
func (p *GoPool) CtxGo(ctx context.Context, f func()) {
	select {
	case p.tasks <- task{ctx: ctx, f: f}:
	default:
		// task queue is full, fallback to use go directly
		go p.runTask(ctx, f)
		return
	}
	// luckily ... it's true when there are many idle workers
	if len(p.tasks) == 0 {
		return
	}
	// all workers are busy, create a new one
	go p.runWorker()
}

// SetPanicHandler sets a function for handling panic.
//
// Panic handler takes two args, `ctx` and `r`.
// `ctx` is the one provided when calling CtxGo, and `r` is returned by recover()
//
// By default, GoPool uses slog to record the err and stack.
//
// It's recommended to set your own handler.
func (p *GoPool) SetPanicHandler(handler PanicHandler) {
	p.panicHandler = handler
}

func (p *GoPool) CurrentWorkers() int {
	return int(atomic.LoadInt32(&p.workers))
}

func (p *GoPool) runTask(ctx context.Context, f func()) {
	defer func(ctx context.Context, p *GoPool) {
		if r := recover(); r != nil {
			if p.panicHandler != nil {
				p.panicHandler(ctx, r)
			} else {
				defaultPanicHandler(ctx, r)
			}
		}
	}(ctx, p)
	f()
}

func (p *GoPool) runWorker() {
	id := atomic.AddInt32(&p.workers, 1)
	defer atomic.AddInt32(&p.workers, -1)

	// worker numbers exceeds maxIdle, drain task and  exit without waiting
	if id > p.maxIdle {
		for {
			select {
			case t := <-p.tasks:
				p.runTask(t.ctx, t.f)
			default:
				return
			}
		}
	}

	tptr := &p.unixMilli
	maxAge := p.maxAge.Milliseconds()
	createdAt := time.Now().UnixMilli()
	for t := range p.tasks {
		p.runTask(t.ctx, t.f)
		// start ticker if it's NOT running
		now := atomic.LoadInt64(tptr)
		if now == 0 {
			now = time.Now().UnixMilli()
			if atomic.CompareAndSwapInt64(tptr, 0, now) {
				go p.runTicker()
			}
		}
		// destroy the worker if maxAge is reached
		if now-createdAt > maxAge {
			return
		}
	}
}

func (p *GoPool) runTicker() {
	// Mark zero to trigger the ticker when we have active workers.
	defer atomic.StoreInt64(&p.unixMilli, 0)

	// If maxAge is 1 min, it updates `unixMilli` and sends 1000 noop tasks per minute.
	// As a result, workers may take longer time to exit, which is expected.
	// We set a minimum interval to avoid performance issues.
	const minInterval = 10 * time.Millisecond
	d := max(p.maxAge/1000, minInterval)
	n0 := int(max(1, p.maxIdle/int32(p.maxAge/d)))

	t := time.NewTicker(d)
	defer t.Stop()
	for now := range t.C {
		x := p.CurrentWorkers()
		if x == 0 {
			return
		}
		atomic.StoreInt64(&p.unixMilli, now.UnixMilli())
		n := max(1, min(n0, x, cap(p.tasks)-len(p.tasks)))
		for i := 0; i < n; i++ {
			p.tasks <- noopTask
		}
	}
}
