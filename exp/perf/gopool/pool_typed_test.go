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
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

type incInt32Data struct {
	wg *sync.WaitGroup
	n  *int32
}

func testIncInt32(_ context.Context, arg incInt32Data) {
	defer arg.wg.Done()
	atomic.AddInt32(arg.n, 1)
}

func TestTypedPool(t *testing.T) {
	p := NewTypedPool(testIncInt32, &Config{AdhocWorkerLimit: 100})
	testWithTypedPool(t, p)
}

func TestTypedPoolWithPermanentWorkers(t *testing.T) {
	p := NewTypedPool(testIncInt32, &Config{
		PermanentWorkerNum: 100,
		AdhocWorkerLimit:   100,
	})
	testWithTypedPool(t, p)
}

func testWithTypedPool(t *testing.T, p *TypedPool[incInt32Data]) {
	var n int32
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		p.Go(incInt32Data{&wg, &n})
	}
	wg.Wait()
	if n != 2000 {
		t.Error(n)
	}
}

func TestTypedPoolPanic(t *testing.T) {
	p := NewTypedPool(func(_ context.Context, arg incInt32Data) {
		defer arg.wg.Done()
		panic("test panic")
	}, &Config{AdhocWorkerLimit: 100})

	var n int32
	var wg sync.WaitGroup
	wg.Add(1)
	p.Go(incInt32Data{&wg, &n})
	wg.Wait()
}

func BenchmarkTypedPool(b *testing.B) {
	p := NewTypedPool(func(_ context.Context, wg *sync.WaitGroup) {
		testFunc()
		wg.Done()
	}, &Config{
		AdhocWorkerLimit: runtime.GOMAXPROCS(0),
	})
	benchmarkWithTypedPool(b, p)
}

func BenchmarkTypedPoolWithPermanentWorkers(b *testing.B) {
	p := NewTypedPool(func(_ context.Context, wg *sync.WaitGroup) {
		testFunc()
		wg.Done()
	}, &Config{
		PermanentWorkerNum: runtime.GOMAXPROCS(0),
		AdhocWorkerLimit:   runtime.GOMAXPROCS(0),
	})
	benchmarkWithTypedPool(b, p)
}

func benchmarkWithTypedPool(b *testing.B, p *TypedPool[*sync.WaitGroup]) {
	var wg sync.WaitGroup
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			p.Go(&wg)
		}
		wg.Wait()
	}
}
