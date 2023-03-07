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
	"sync"
	"testing"
	"time"
)

func TestDefaultPool(t *testing.T) {
	var mu sync.Mutex
	var x int
	var ch = make(chan struct{}, 10)
	Go(func() {
		mu.Lock()
		x++
		mu.Unlock()
		ch <- struct{}{}
	})

	var adhocWorkerCnt int32
	CtxGo(context.Background(), func() {
		adhocWorkerCnt = Default().AdhocWorkerCount()

		mu.Lock()
		x++
		mu.Unlock()
		ch <- struct{}{}
	})

	for i := 0; i < 2; i++ {
		<-ch
	}
	if adhocWorkerCnt == 0 {
		t.Errorf("adhocWorkerCnt == 0")
	}
	time.Sleep(100 * time.Millisecond)
	if n := Default().AdhocWorkerCount(); n != 0 {
		t.Errorf("defualtPool adhoc worker count, want 0, got %d", n)
	}
}

var registerTestPoolOnce sync.Once

func TestRegister(t *testing.T) {
	p := NewPool(&Config{Name: "testPool"})

	// Use sync.Once to avoid error when run with argument -count=N where N > 1.
	registerTestPoolOnce.Do(func() {
		err := Register(p)
		if err != nil {
			t.Error(err)
		}
	})

	p = Get("testPool")
	if p == nil {
		t.Error("Get did not return registered pool")
	}
}
