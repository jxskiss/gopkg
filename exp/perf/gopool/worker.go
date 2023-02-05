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

import "context"

type taskRunner func(p *Pool, t *task)

func runPermanentWorker(p *Pool, runner taskRunner) {
	for {
		select {
		case t := <-p.taskCh:
			runner(p, t)

			// Drain pending tasks.
			for {
				t = p.taskList.pop()
				if t == nil {
					break
				}
				runner(p, t)
			}
		}
	}
}

func runAdhocWorker(p *Pool, runner taskRunner) {
	go func() {
		for {
			t := p.taskList.pop()
			if t == nil {
				p.decWorkerCount()
				return
			}
			runner(p, t)
		}
	}()
}

func funcTaskRunner(p *Pool, t *task) {
	defer func() {
		if r := recover(); r != nil {
			p.config.PanicHandler(t.ctx, r)
		}
	}()
	t.arg.(func())()
	t.Recycle()
}

func newSpecificTaskRunner[T any](handler func(context.Context, T)) taskRunner {
	return func(p *Pool, t *task) {
		defer func() {
			if r := recover(); r != nil {
				p.config.PanicHandler(t.ctx, r)
			}
		}()
		handler(t.ctx, t.arg.(T))
		t.Recycle()
	}
}
