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
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

const benchmarkTimes = 10000

func DoCopyStack(a, b int) int {
	if b < 100 {
		return DoCopyStack(0, b+1)
	}
	return 0
}

func testFunc() {
	DoCopyStack(0, 0)
}

func TestPool(t *testing.T) {
	p := NewPool(&Config{AdhocWorkerLimit: 100})
	testWithPool(t, p)
}

func TestPoolWithPermanentWorkers(t *testing.T) {
	p := NewPool(&Config{
		PermanentWorkerNum: 100,
		AdhocWorkerLimit:   100,
	})
	testWithPool(t, p)
}

func testWithPool(t *testing.T, p *Pool) {
	var n int32
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		p.Go(func() {
			defer wg.Done()
			atomic.AddInt32(&n, 1)
		})
	}
	wg.Wait()
	if n != 2000 {
		t.Error(n)
	}
}

func TestPoolPanic(t *testing.T) {
	p := NewPool(&Config{AdhocWorkerLimit: 100})
	var wg sync.WaitGroup
	wg.Add(1)
	p.Go(func() {
		defer wg.Done()
		panic("test panic")
	})
	wg.Wait()
}

func BenchmarkDefaultPool(b *testing.B) {
	p := NewPool(&Config{
		AdhocWorkerLimit: runtime.GOMAXPROCS(0),
	})
	benchmarkWithPool(b, p)
}

func BenchmarkPoolWithPermanentWorkers(b *testing.B) {
	p := NewPool(&Config{
		PermanentWorkerNum: runtime.GOMAXPROCS(0),
		AdhocWorkerLimit:   runtime.GOMAXPROCS(0),
	})
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
