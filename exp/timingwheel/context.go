package timingwheel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// WithTimeout returns a context value with time.Now().Add(timeout) as deadline.
//
// If the provided timeout is greater than 80 seconds, or the parent context
// already has deadline setup, this function simply calls
// `context.WithDeadline(parent, time.Now().Add(timeout))`.
//
// Otherwise, it uses a shared timing wheel in underlying to achieve high
// performance when managing massive contexts with deadlines. The timing
// wheel's tick interval is 10 milliseconds, thus the precision of the
// timeout is roughly ±10 milliseconds. When shared timing wheel
// implementation is used, the returned context does not support canceling,
// and does not propagate.
//
// NOTE: this context implementation is not intended to replace usage of the
// standard context package. Generally, the standard context package is
// preferred over this in most use cases, only use this where you are managing
// truly massive contexts, and super high performance really outweighted
// functionality and compatibility provided by the standard context package.
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > ctxMaxTimeout {
		if parent == nil {
			parent = context.Background()
		}
		return context.WithDeadline(parent, time.Now().Add(timeout))
	}

	// The context may already has deadline setup.
	if parent != nil {
		if _, ok := parent.Deadline(); ok {
			return context.WithDeadline(parent, time.Now().Add(timeout))
		}
	}

	// lazy initialization
	initCtxWheel()

	deadline := time.Now().Add(timeout)
	done := ctxWheel.getOrCreateTimeoutChan(timeout)
	ctx := &timeoutCtx{
		parent:   parent,
		deadline: deadline,
		done:     done,
	}
	return ctx, emptyCancel
}

func initCtxWheel() {
	ctxOnce.Do(func() {
		ctxWheel = &contextWheel{}
		go ctxWheel.run()
	})
}

const (
	ctxWheelSize    = 8192 // 81.92 seconds
	ctxWheelMask    = ctxWheelSize - 1
	ctxTickInterval = 10 * time.Millisecond
	ctxMaxTimeout   = 80 * time.Second
)

var (
	ctxOnce     sync.Once
	ctxWheel    *contextWheel
	emptyCancel = func() {}
)

type contextWheel struct {
	jiffies int64
	buckets [ctxWheelSize]unsafe.Pointer
}

func (p *contextWheel) run() {
	var jiffies int64
	var ticker = time.NewTicker(ctxTickInterval)
	for range ticker.C {
		idx := jiffies & ctxWheelMask
		chptr := atomic.SwapPointer(&p.buckets[idx], nil)
		if chptr != nil {
			done := *(*chan struct{})(chptr)
			close(done)
		}
		jiffies += 1
		atomic.StoreInt64(&p.jiffies, jiffies)
	}
}

func (p *contextWheel) getOrCreateTimeoutChan(timeout time.Duration) <-chan struct{} {
	expires := int64(timeout / ctxTickInterval)
	if expires > 0 {
		expires -= 1
	}
	for {
		jiffies := atomic.LoadInt64(&p.jiffies)
		idx := (jiffies + expires) & ctxWheelMask
		ch := atomic.LoadPointer(&p.buckets[idx])
		if ch != nil {
			return *(*chan struct{})(ch)
		}

		newCh := make(chan struct{})
		swapped := atomic.CompareAndSwapPointer(&p.buckets[idx], nil, unsafe.Pointer(&newCh))
		if swapped {
			return newCh
		}
	}
}

// timeoutCtx implements the context.Context interface.
type timeoutCtx struct {
	parent   context.Context
	deadline time.Time
	done     <-chan struct{}
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
	if c.parent == nil {
		return nil
	}
	return c.parent.Value(key)
}
