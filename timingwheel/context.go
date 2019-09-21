package wheel

import (
	"context"
	"github.com/jxskiss/gopkg/fastrand"
	"sync"
	"time"
)

const (
	cacheLineSize = 64

	ctxTickInterval = defaultWheelInterval
	ctxMaxTimeout   = 32 * time.Second
	ctxWheelSize    = ctxMaxTimeout / ctxTickInterval
	ctxBucketShard  = 64
)

type bucket struct {
	ch chan struct{}
	//_padding [cacheLineSize - unsafe.Sizeof(make(chan struct{}))%cacheLineSize]byte
}

var (
	ctxInitOnce  sync.Once
	timeoutWheel = struct {
		mu  sync.RWMutex
		cs  [ctxWheelSize][ctxBucketShard]bucket
		pos int
	}{}
)

func initTimeoutWheel() {
	for i := ctxWheelSize - 1; i >= 0; i-- {
		for j := ctxBucketShard - 1; j >= 0; j-- {
			timeoutWheel.cs[i][j].ch = make(chan struct{})
		}
	}

	onTick := func() {
		var newBuckets [ctxBucketShard]bucket
		for i := 0; i < ctxBucketShard; i++ {
			newBuckets[i].ch = make(chan struct{})
		}

		w := &timeoutWheel
		w.mu.Lock()
		oldBuckets := w.cs[w.pos]
		w.cs[w.pos] = newBuckets
		w.pos = (w.pos + 1) % len(w.cs)
		w.mu.Unlock()

		for i := 0; i < ctxBucketShard; i++ {
			close(oldBuckets[i].ch)
		}
	}
	defaultWheel().TickFunc(ctxTickInterval, onTick)
}

func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout >= ctxMaxTimeout {
		return context.WithTimeout(parent, timeout)
	}

	ctxInitOnce.Do(initTimeoutWheel)

	bucketIdx := fastrand.Uint32() % ctxBucketShard
	wheelIdx := int(timeout / ctxTickInterval)
	if 0 < wheelIdx {
		wheelIdx--
	}

	w := &timeoutWheel
	w.mu.RLock()
	wheelIdx = (w.pos + wheelIdx) % len(w.cs)
	done := w.cs[wheelIdx][bucketIdx].ch
	w.mu.RUnlock()

	ctx := &timeoutCtx{
		done:     done,
		deadline: time.Now().Add(timeout),
	}
	cancel := func() {}
	return ctx, cancel
}

type timeoutCtx struct {
	done     <-chan struct{}
	deadline time.Time
}

func (c *timeoutCtx) Deadline() (deadline time.Time, ok bool) {
	return c.deadline, true
}

func (c *timeoutCtx) Done() <-chan struct{} {
	return c.done
}

func (c *timeoutCtx) Err() error {
	if _, ok := <-c.done; ok {
		return context.DeadlineExceeded
	}
	return nil
}

func (c *timeoutCtx) Value(key interface{}) interface{} {
	return nil
}
