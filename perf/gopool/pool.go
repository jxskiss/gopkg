// Copyright 2021 ByteDance Inc.
// Copyright 2023 Shawn Wang <jxskiss@126.com>.
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
	"math"
	"sync"
	"sync/atomic"
)

// Pool manages a goroutine pool and tasks for better performance,
// it reuses goroutines and limits the number of goroutines.
type Pool = TypedPool[func()]

// NewPool creates a new pool with the config.
func NewPool(config *Config) *Pool {
	runner := funcTaskRunner
	p := &TypedPool[func()]{}
	p.init(config, runner)
	return p
}

// TypedPool is a task-specific pool.
// A TypedPool is like pool, but it executes a handler to process values
// of a specific type.
// Compared to Pool, it helps to reduce unnecessary memory allocation of
// closures when submitting tasks.
type TypedPool[T any] struct {
	internalPool
}

// NewTypedPool creates a new task-specific pool with given handler and config.
func NewTypedPool[T any](config *Config, handler func(context.Context, T)) *TypedPool[T] {
	runner := newTypedTaskRunner(handler)
	p := &TypedPool[T]{}
	p.init(config, runner)
	return p
}

// Go submits a task to the pool.
func (p *TypedPool[T]) Go(arg T) {
	p.submit(context.Background(), arg)
}

// CtxGo submits a task to the pool, it's preferred over Go.
func (p *TypedPool[T]) CtxGo(ctx context.Context, arg T) {
	p.submit(ctx, arg)
}

type internalPool struct {
	config *Config
	runner taskRunner

	// taskCh sends tasks to permanent workers.
	taskCh chan *task

	// mu protects adhocState and taskList.
	// adhocState:
	// - higher 32 bits is adhocLimit, max number of adhoc workers that can run simultaneously
	// - lower 32 bits is adhocCount, the number of currently running adhoc workers
	mu         sync.Mutex
	adhocState int64
	taskList   taskList
}

func (p *internalPool) init(config *Config, runner taskRunner) {
	config.checkAndSetDefaults()
	p.config = config
	p.runner = runner
	p.SetAdhocWorkerLimit(config.AdhocWorkerLimit)
	p.startPermanentWorkers()
}

// Name returns the name of a pool.
func (p *internalPool) Name() string {
	return p.config.Name
}

// SetAdhocWorkerLimit changes the limit of adhoc workers.
// 0 or negative value means no limit.
func (p *internalPool) SetAdhocWorkerLimit(limit int) {
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

func (p *internalPool) submit(ctx context.Context, arg any) {
	t := newTask()
	t.ctx = ctx
	t.arg = arg

	// Try permanent worker first.
	select {
	case p.taskCh <- t:
		return
	default:
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
	if wCnt == 0 || (tCnt >= p.config.ScaleThreshold && wCnt < wLimit) {
		p.incAdhocWorkerCount()
		p.mu.Unlock()
		go p.adhocWorker(t)
	} else {
		p.taskList.add(t)
		p.mu.Unlock()
	}
}

// AdhocWorkerLimit returns the current limit of adhoc workers.
func (p *internalPool) AdhocWorkerLimit() int32 {
	limit, _ := p.getAdhocState()
	return limit
}

// AdhocWorkerCount returns the number of running adhoc workers.
func (p *internalPool) AdhocWorkerCount() int32 {
	_, count := p.getAdhocState()
	return count
}

func (p *internalPool) getAdhocState() (limit, count int32) {
	x := atomic.LoadInt64(&p.adhocState)
	return int32(x >> 32), int32((x << 32) >> 32)
}

func (p *internalPool) incAdhocWorkerCount() {
	atomic.AddInt64(&p.adhocState, 1)
}

func (p *internalPool) decAdhocWorkerCount() {
	atomic.AddInt64(&p.adhocState, -1)
}

// PermanentWorkerCount returns the number of permanent workers.
func (p *internalPool) PermanentWorkerCount() int32 {
	return int32(p.config.PermanentWorkerNum)
}

func (p *internalPool) startPermanentWorkers() {
	if p.config.PermanentWorkerNum <= 0 {
		return
	}
	p.taskCh = make(chan *task)
	for i := 0; i < p.config.PermanentWorkerNum; i++ {
		go p.permanentWorker()
	}
}

func (p *internalPool) permanentWorker() {
	for {
		select {
		case t := <-p.taskCh:
			p.runner(p, t)

			// Drain pending tasks.
			for {
				p.mu.Lock()
				t = p.taskList.pop()
				p.mu.Unlock()
				if t == nil {
					break
				}
				p.runner(p, t)
			}
		}
	}
}

func (p *internalPool) adhocWorker(t *task) {
	p.runner(p, t)
	for {
		p.mu.Lock()
		t = p.taskList.pop()
		if t == nil {
			p.decAdhocWorkerCount()
			p.mu.Unlock()
			return
		}
		p.mu.Unlock()
		p.runner(p, t)
	}
}
