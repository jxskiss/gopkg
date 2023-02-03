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

func runPermanentWorker(p *Pool) {
	for {
		select {
		case t := <-p.taskCh:
			doTask(p, t)

			// Drain pending tasks.
			for {
				t = p.taskList.pop()
				if t == nil {
					break
				}
				doTask(p, t)
			}
		}
	}
}

func runAdhocWorker(p *Pool) {
	go func() {
		for {
			t := p.taskList.pop()
			if t == nil {
				p.decWorkerCount()
				return
			}
			doTask(p, t)
		}
	}()
}

func doTask(p *Pool, t *task) {
	defer func() {
		if r := recover(); r != nil {
			p.config.PanicHandler(t.ctx, r)
		}
	}()
	t.f()
	t.Recycle()
}
