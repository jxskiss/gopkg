package gopool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestUnboundedPool(t *testing.T) {
	cfg := NewConfig().SetUnbounded(100, time.Minute)
	p := NewPool(cfg)

	var n int32
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		p.Go(func() {
			defer wg.Done()
			atomic.AddInt32(&n, 1)
			time.Sleep(100 * time.Millisecond)
		})
	}
	wg.Wait()
	if n != 2000 {
		t.Error(n)
	}
	time.Sleep(10 * time.Millisecond)
	if x := p.AdhocWorkerCount(); x != 100 {
		t.Errorf("adhoc worker count, want 100, got %d", x)
	}
}

func TestUnboundedPoolPanic(t *testing.T) {
	p := NewPool(NewConfig().SetUnbounded(100, time.Minute))
	var wg sync.WaitGroup
	wg.Add(1)
	p.Go(func() {
		defer wg.Done()
		panic("test panic")
	})
	wg.Wait()
}

func BenchmarkUnboundedPool(b *testing.B) {
	p := NewPool(NewConfig().SetUnbounded(1000, time.Minute))
	benchmarkWithPool(b, p)
}

func BenchmarkUnboundedPoolParallel(b *testing.B) {
	p := NewPool(NewConfig().SetUnbounded(1000, time.Minute))
	benchmarkWithPoolParallel(b, p)
}

func BenchmarkUnboundedTypedPool(b *testing.B) {
	cfg := NewConfig().SetUnbounded(1000, time.Minute)
	p := NewTypedPool(cfg, func(_ context.Context, wg *sync.WaitGroup) {
		testFunc()
		wg.Done()
	})
	benchmarkWithTypedPool(b, p)
}

func BenchmarkUnboundedTypedPoolParallel(b *testing.B) {
	cfg := NewConfig().SetUnbounded(1000, time.Minute)
	p := NewTypedPool(cfg, func(_ context.Context, wg *sync.WaitGroup) {
		testFunc()
		wg.Done()
	})
	benchmarkWithTypedPoolParallel(b, p)
}
