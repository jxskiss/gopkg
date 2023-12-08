package syncx

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type spinLock uintptr

// NewSpinLock creates a new spin-lock.
// A spin-lock calls runtime.Gosched when it failed acquiring the lock
// by compare-and-swap operation, then retry until it succeeds.
func NewSpinLock() sync.Locker {
	return new(spinLock)
}

// Lock acquires the lock.
func (p *spinLock) Lock() {
	if !atomic.CompareAndSwapUintptr((*uintptr)(p), 0, 1) {
		// Outlined slow-path to allow inlining of the fast-path.
		p.lockSlowPath()
	}
}

func (p *spinLock) lockSlowPath() {
	const maxBackoff = 8
	backoff := 1
	for {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if atomic.CompareAndSwapUintptr((*uintptr)(p), 0, 1) {
			break
		}
		backoff <<= 1
		if backoff > maxBackoff {
			backoff = 1
		}
	}
}

// Unlock releases the lock.
func (p *spinLock) Unlock() {
	atomic.StoreUintptr((*uintptr)(p), 0)
}
