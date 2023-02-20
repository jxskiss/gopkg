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

func TestRegister(t *testing.T) {
	p := NewPool(&Config{Name: "testPool"})
	err := Register(p)
	if err != nil {
		t.Error(err)
	}
	p = Get("testPool")
	if p == nil {
		t.Error("Get did not return registered pool")
	}
}
