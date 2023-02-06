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
)

// SpecificPool is a task-specific pool.
// A SpecificPool is like pool, but it executes a handler to process values
// of a specific type.
// Compared to Pool, it helps to reduce unnecessary memory allocation of
// closures when submitting tasks.
type SpecificPool[T any] struct {
	pool    *Pool
	handler func(context.Context, T)
	runner  taskRunner
}

// NewSpecificPool creates a new task-specific pool with given handler and config.
func NewSpecificPool[T any](handler func(context.Context, T), config *Config) *SpecificPool[T] {
	config.checkAndSetDefaults()
	p := &SpecificPool[T]{
		pool: &Pool{
			config:     config,
			adhocLimit: getAdhocWorkerLimit(config.AdhocWorkerLimit),
		},
		handler: handler,
		runner:  newSpecificTaskRunner(handler),
	}
	p.pool.spawnPermanentWorkers(p.runner)
	return p
}

// Name returns the name of a pool.
func (p *SpecificPool[_]) Name() string {
	return p.pool.config.Name
}

// SetAdhocWorkerLimit changes the limit of adhoc workers.
// 0 or negative value means no limit.
func (p *SpecificPool[_]) SetAdhocWorkerLimit(limit int) {
	p.pool.SetAdhocWorkerLimit(limit)
}

// Go submits a task to the pool.
func (p *SpecificPool[T]) Go(arg T) {
	p.CtxGo(context.Background(), arg)
}

// CtxGo submits a task to the pool, it's preferred over Go.
func (p *SpecificPool[T]) CtxGo(ctx context.Context, arg T) {
	p.pool.submit(ctx, arg, p.runner)
}

// AdhocWorkerLimit returns the current limit of adhoc workers.
func (p *SpecificPool[_]) AdhocWorkerLimit() int32 {
	return p.pool.AdhocWorkerLimit()
}

// AdhocWorkerCount returns the number of running adhoc workers.
func (p *SpecificPool[_]) AdhocWorkerCount() int32 {
	return p.pool.AdhocWorkerCount()
}

// PermanentWorkerCount returns the number of permanent workers.
func (p *SpecificPool[_]) PermanentWorkerCount() int32 {
	return p.pool.PermanentWorkerCount()
}
