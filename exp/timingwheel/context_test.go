package timingwheel

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

func waitCtxTimeout(ctx context.Context, timeout time.Duration) {
	before := time.Now()
	<-ctx.Done()
	log.Printf("%v -> %v", timeout, time.Since(before))
}

func TestWithTimeout(t *testing.T) {
	timeouts := []time.Duration{
		10 * time.Millisecond,
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
		5 * time.Second,
		10 * time.Second,
		20 * time.Second,
		50 * time.Second,
		50 * time.Second,
		80 * time.Second,
		90 * time.Second,
		100 * time.Second,
		100 * time.Second,
	}

	wg := sync.WaitGroup{}
	for _, d := range timeouts {
		d := d
		ctx, _ := WithTimeout(nil, d)
		wg.Add(1)
		go func() {
			waitCtxTimeout(ctx, d)
			wg.Done()
		}()
	}

	time.Sleep(time.Second)
	ctx1, _ := WithTimeout(nil, time.Second)
	wg.Add(1)
	go func() {
		waitCtxTimeout(ctx1, time.Second)
		wg.Done()
	}()

	time.Sleep(time.Second)
	ctx2, _ := WithTimeout(nil, 4*time.Second)
	wg.Add(1)
	go func() {
		waitCtxTimeout(ctx2, 4*time.Second)
		wg.Done()
	}()

	wg.Wait()
}
