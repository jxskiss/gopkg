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
	"sync/atomic"
)

type Pool struct {

	// The name of the pool
	name string

	// Configuration information
	config *Config

	// Limit of adhoc workers that can run simultaneously
	adhocLimit int32

	// Number of running adhoc workers
	adhocCount int32

	taskCh   chan *task
	taskList taskList
}

// NewPool creates a new pool with the given name and config.
func NewPool(name string, config *Config) *Pool {
	config.checkAndSetDefaults()
	p := &Pool{
		name:       name,
		config:     config,
		adhocLimit: getAdhocWorkerLimit(config.AdhocWorkerLimit),
	}
	p.spawnPermanentWorkers()
	return p
}

// Name returns the name of a pool.
func (p *Pool) Name() string {
	return p.name
}

// SetAdhocWorkerLimit changes the limit of adhoc workers.
// 0 or negative value means no limit.
func (p *Pool) SetAdhocWorkerLimit(limit int) {
	atomic.StoreInt32(&p.adhocLimit, getAdhocWorkerLimit(limit))
}

// Go submits a function to the pool.
func (p *Pool) Go(f func()) {
	p.CtxGo(context.Background(), f)
}

// CtxGo submits a function to the pool, it's preferred over Go.
func (p *Pool) CtxGo(ctx context.Context, f func()) {
	t := newTask()
	t.ctx = ctx
	t.f = f

	// Try permanent worker first.
	select {
	case p.taskCh <- t:
		return
	default:
	}

	// No permanent worker available or all workers are busy, submit to task list.
	tCnt := p.taskList.add(t)

	// The following two conditions are met:
	//   1. the number of tasks is greater than the threshold.
	//   2. The current number of workers is less than the upper limit p.cap.
	//
	// Or there are currently no workers.
	limit := p.AdhocWorkerLimit()
	wCnt := p.AdhocWorkerCount()
	if (tCnt >= p.config.ScaleThreshold && wCnt < limit) || wCnt == 0 {
		p.incWorkerCount()
		runAdhocWorker(p)
	}
}

// AdhocWorkerLimit returns the current limit of adhoc workers.
func (p *Pool) AdhocWorkerLimit() int32 {
	return atomic.LoadInt32(&p.adhocLimit)
}

// AdhocWorkerCount returns the number of running adhoc workers.
func (p *Pool) AdhocWorkerCount() int32 {
	return atomic.LoadInt32(&p.adhocCount)
}

// PermanentWorkerCount returns the number of permanent workers.
func (p *Pool) PermanentWorkerCount() int32 {
	return int32(p.config.PermanentWorkerNum)
}

func (p *Pool) incWorkerCount() {
	atomic.AddInt32(&p.adhocCount, 1)
}

func (p *Pool) decWorkerCount() {
	atomic.AddInt32(&p.adhocCount, -1)
}

func (p *Pool) spawnPermanentWorkers() {
	if p.config.PermanentWorkerNum <= 0 {
		return
	}
	p.taskCh = make(chan *task)
	for i := 0; i < p.config.PermanentWorkerNum; i++ {
		go runPermanentWorker(p)
	}
}
