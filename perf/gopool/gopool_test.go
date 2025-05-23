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
	"bytes"
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/jxskiss/gopkg/v2/internal"
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
		adhocWorkerCnt = defaultPool.AdhocWorkerCount()

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
	if n := defaultPool.AdhocWorkerCount(); n != 0 {
		t.Errorf("defualtPool adhoc worker count, want 0, got %d", n)
	}
}

func TestDefaultPanicHandler(t *testing.T) {
	f1 := func() {
		log.Println("f1")
	}
	f2 := func() {
		f1()
		panic("panic f2")
	}

	var buf internal.LockedBuffer
	log.SetOutput(&buf)

	Go(f2)
	time.Sleep(100 * time.Millisecond)

	logStr := buf.String()
	if !bytes.Contains([]byte(logStr), []byte("[ERROR] gopool: catch panic: panic f2")) {
		t.Errorf("log output does not contain panic message")
	}
	if !bytes.Contains([]byte(logStr), []byte("perf/gopool.TestDefaultPanicHandler.func2:")) {
		t.Errorf("log output does not contain panic location")
	}
}
