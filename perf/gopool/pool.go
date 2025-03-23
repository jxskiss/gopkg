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
	config *Config

	boundedPool   *boundedPool
	unboundedPool *unboundedPool
	submitFunc    func(context.Context, any)
}

// NewTypedPool creates a new task-specific pool with given handler and config.
func NewTypedPool[T any](config *Config, handler func(context.Context, T)) *TypedPool[T] {
	runner := newTypedTaskRunner(handler)
	p := &TypedPool[T]{}
	p.init(config, runner)
	return p
}

func (p *TypedPool[T]) init(config *Config, runner taskRunner) {
	if config.isUnbounded() {
		p.unboundedPool = newUnboundedPool(config, runner)
		p.submitFunc = p.unboundedPool.submitTask
	} else {
		p.boundedPool = newBoundedPool(config, runner)
		p.submitFunc = p.boundedPool.submitTask
	}
}

// Go submits a task to the pool.
func (p *TypedPool[T]) Go(arg T) {
	p.submitFunc(context.Background(), arg)
}

// CtxGo submits a task to the pool, it's preferred over Go.
func (p *TypedPool[T]) CtxGo(ctx context.Context, arg T) {
	p.submitFunc(ctx, arg)
}

// Name returns the name of a pool.
func (p *TypedPool[T]) Name() string {
	return p.config.Name
}

// SetAdhocWorkerLimit changes the limit of adhoc workers.
// 0 or negative value means no limit.
// For an unbounded pool, calling this method is a no-op.
func (p *TypedPool[T]) SetAdhocWorkerLimit(limit int) {
	if p.boundedPool != nil {
		p.boundedPool.setAdhocWorkerLimit(limit)
	}
}

// AdhocWorkerLimit returns the current limit of adhoc workers.
// For an unbounded pool, it always returns math.MaxInt32.
func (p *TypedPool[T]) AdhocWorkerLimit() int {
	if p.boundedPool != nil {
		limit, _ := p.boundedPool.getAdhocState()
		return int(limit)
	}
	return math.MaxInt32
}

// AdhocWorkerCount returns the number of running adhoc workers.
// For an unbounded pool, it returns the number of all running workers.
func (p *TypedPool[T]) AdhocWorkerCount() int {
	if p.boundedPool != nil {
		_, count := p.boundedPool.getAdhocState()
		return int(count)
	}
	return p.unboundedPool.workersCount()
}
