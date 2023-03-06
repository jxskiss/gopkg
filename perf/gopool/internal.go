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
	"sync/atomic"
)

type internalPool struct {
	config *Config
	runner taskRunner

	// Limit of adhoc workers that can run simultaneously
	adhocLimit int32

	// Number of running adhoc workers
	adhocCount int32

	taskCh   chan *task
	taskList taskList
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
	atomic.StoreInt32(&p.adhocLimit, int32(limit))
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

	// No permanent worker available or all workers are busy, submit to task list.
	tCnt := p.taskList.add(t)

	// The following two conditions are met:
	//   1. the number of tasks is greater than the threshold.
	//   2. The current number of workers is less than the upper limit p.cap.
	//
	// Or there are currently no workers.
	wCnt := p.AdhocWorkerCount()
	if wCnt == 0 || (tCnt >= p.config.ScaleThreshold && wCnt < p.AdhocWorkerLimit()) {
		p.runAdhocWorker()
	}
}

// AdhocWorkerLimit returns the current limit of adhoc workers.
func (p *internalPool) AdhocWorkerLimit() int32 {
	return atomic.LoadInt32(&p.adhocLimit)
}

// AdhocWorkerCount returns the number of running adhoc workers.
func (p *internalPool) AdhocWorkerCount() int32 {
	return atomic.LoadInt32(&p.adhocCount)
}

// PermanentWorkerCount returns the number of permanent workers.
func (p *internalPool) PermanentWorkerCount() int32 {
	return int32(p.config.PermanentWorkerNum)
}

func (p *internalPool) incWorkerCount() {
	atomic.AddInt32(&p.adhocCount, 1)
}

func (p *internalPool) decWorkerCount() {
	atomic.AddInt32(&p.adhocCount, -1)
}

func (p *internalPool) startPermanentWorkers() {
	if p.config.PermanentWorkerNum <= 0 {
		return
	}
	p.taskCh = make(chan *task)
	for i := 0; i < p.config.PermanentWorkerNum; i++ {
		go p.runPermanentWorker()
	}
}
