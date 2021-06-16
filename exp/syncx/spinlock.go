package syncx

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type spinLock uint32

// NewSpinLock creates a new spin lock.
// A spin lock calls runtime.Gosched when it failed acquiring the lock,
// then try again until it success.
func NewSpinLock() sync.Locker {
	return new(spinLock)
}

// Lock acquires the lock.
func (p *spinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(p), 0, 1) {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		backoff <<= 1
	}
}

// Unlock releases the lock.
func (p *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(p), 0)
}
