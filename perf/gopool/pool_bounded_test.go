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
	"time"
)

const benchmarkTimes = 1000

func DoCopyStack(a, b int) int {
	if b < 100 {
		return DoCopyStack(0, b+1)
	}
	return 0
}

func testFunc() {
	DoCopyStack(0, 0)
}

func TestBoundedPool(t *testing.T) {
	cfg := NewConfig().SetBounded(1, 0, 100)
	p := NewPool(cfg)
	testWithBoundedPool(t, p, 100)
}

func TestBoundedPoolWithPermanentWorkers(t *testing.T) {
	cfg := NewConfig().SetBounded(0, 100, 100)
	p := NewPool(cfg)
	testWithBoundedPool(t, p, 100)
}

func testWithBoundedPool(t *testing.T, p *Pool, adhocLimit int) {
	var n int32
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		p.Go(func() {
			defer wg.Done()
			atomic.AddInt32(&n, 1)
			if x := p.AdhocWorkerCount(); x > adhocLimit {
				t.Errorf("adhoc worker count, want <= %d, got %d", adhocLimit, x)
			}
		})
	}
	wg.Wait()
	if n != 2000 {
		t.Error(n)
	}
	time.Sleep(100 * time.Millisecond)
	if x := p.AdhocWorkerCount(); x != 0 {
		t.Errorf("adhoc worker count, want 0, got %d", x)
	}
}

func TestBoundedPoolPanic(t *testing.T) {
	p := NewPool(NewConfig().SetBounded(0, 0, 100))
	var wg sync.WaitGroup
	wg.Add(1)
	p.Go(func() {
		defer wg.Done()
		panic("test panic")
	})
	wg.Wait()
}

func BenchmarkBoundedPool(b *testing.B) {
	p := NewPool(NewConfig().SetBounded(0, 0, runtime.GOMAXPROCS(0)))
	benchmarkWithPool(b, p)
}

func BenchmarkBoundedPoolParallel(b *testing.B) {
	p := NewPool(NewConfig().SetBounded(0, 0, runtime.GOMAXPROCS(0)))
	benchmarkWithPoolParallel(b, p)
}

func BenchmarkBoundedPoolWithPermanentWorkers(b *testing.B) {
	p := NewPool(NewConfig().SetBounded(0, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0)))
	benchmarkWithPool(b, p)
}

func benchmarkWithPool(b *testing.B, p *Pool) {
	var wg sync.WaitGroup
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			p.Go(func() {
				testFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func benchmarkWithPoolParallel(b *testing.B, p *Pool) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(benchmarkTimes)
			for j := 0; j < benchmarkTimes; j++ {
				p.Go(func() {
					testFunc()
					wg.Done()
				})
			}
			wg.Wait()
		}
	})
}

func BenchmarkGo(b *testing.B) {
	var wg sync.WaitGroup
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			go func() {
				testFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

type incInt32Data struct {
	wg *sync.WaitGroup
	n  *int32
}

func testIncInt32(_ context.Context, arg incInt32Data) {
	defer arg.wg.Done()
	atomic.AddInt32(arg.n, 1)
}

func TestBoundedTypedPool(t *testing.T) {
	cfg := NewConfig().SetBounded(0, 0, 100)
	p := NewTypedPool(cfg, testIncInt32)
	testWithTypedPool(t, p)
}

func TestBoundedTypedPoolWithPermanentWorkers(t *testing.T) {
	cfg := NewConfig().SetBounded(0, 100, 100)
	p := NewTypedPool(cfg, testIncInt32)
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
	time.Sleep(100 * time.Millisecond)
	if x := p.AdhocWorkerCount(); x != 0 {
		t.Errorf("adhoc worker count, want 0, got %d", x)
	}
}

func TestBoundedTypedPoolPanic(t *testing.T) {
	cfg := NewConfig().SetBounded(0, 0, 100)
	p := NewTypedPool(cfg, func(_ context.Context, arg incInt32Data) {
		defer arg.wg.Done()
		panic("test panic")
	})

	var n int32
	var wg sync.WaitGroup
	wg.Add(1)
	p.Go(incInt32Data{&wg, &n})
	wg.Wait()
}

func BenchmarkBoundedTypedPool(b *testing.B) {
	cfg := NewConfig().SetBounded(0, 0, runtime.GOMAXPROCS(0))
	p := NewTypedPool(cfg, func(_ context.Context, wg *sync.WaitGroup) {
		testFunc()
		wg.Done()
	})
	benchmarkWithTypedPool(b, p)
}

func BenchmarkBoundedTypedPoolParallel(b *testing.B) {
	cfg := NewConfig().SetBounded(0, 0, runtime.GOMAXPROCS(0))
	p := NewTypedPool(cfg, func(_ context.Context, wg *sync.WaitGroup) {
		testFunc()
		wg.Done()
	})
	benchmarkWithTypedPoolParallel(b, p)
}

func BenchmarkBoundedTypedPoolWithPermanentWorkers(b *testing.B) {
	cfg := NewConfig().SetBounded(0, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0))
	p := NewTypedPool(cfg, func(_ context.Context, wg *sync.WaitGroup) {
		testFunc()
		wg.Done()
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

func benchmarkWithTypedPoolParallel(b *testing.B, p *TypedPool[*sync.WaitGroup]) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(benchmarkTimes)
			for j := 0; j < benchmarkTimes; j++ {
				p.Go(&wg)
			}
			wg.Wait()
		}
	})
}

func TestSetAdhocWorkerLimit(t *testing.T) {
	cfg := NewConfig().SetBounded(0, 0, 100)
	pool := NewPool(cfg)
	wg := &sync.WaitGroup{}

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			pool.SetAdhocWorkerLimit(80)
			wg.Done()
		}()
	}
	wg.Wait()
	if x := pool.AdhocWorkerLimit(); x != 80 {
		t.Errorf("adhoc worker limit not match, want 80, got %d", x)
	}
	if x := pool.AdhocWorkerCount(); x != 0 {
		t.Errorf("adhoc worker count not match, want 0, got %d", x)
	}

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			pool.SetAdhocWorkerLimit(100)
			wg.Done()
		}()
	}
	wg.Wait()
	if x := pool.AdhocWorkerLimit(); x != 100 {
		t.Errorf("adhoc worker limit not match, want 100, got %d", x)
	}
	if x := pool.AdhocWorkerCount(); x != 0 {
		t.Errorf("adhoc worker count not match, want 0, got %d", x)
	}

	wg.Add(100)
	for i := 0; i < 100; i++ {
		limit := 80 + 20*(i%2)
		go func() {
			pool.SetAdhocWorkerLimit(limit)
			wg.Done()
		}()
	}
	wg.Wait()
	if x := pool.AdhocWorkerLimit(); x != 100 && x != 80 {
		t.Errorf("adhoc worker limit not match, got %d", x)
	}
	if x := pool.AdhocWorkerCount(); x != 0 {
		t.Errorf("adhoc worker count not match, want 0, got %d", x)
	}
}
