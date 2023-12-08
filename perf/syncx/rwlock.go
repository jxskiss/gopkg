package syncx

import (
	"runtime"
	"sync"

	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

const cacheLineSize = 64

// RWLock holds a group of sharded RWMutex, it gives great performance
// in read-heavy workloads by reducing lock contention, but the performance
// for exclusive Lock is poor.
type RWLock struct {
	shards []rwlockShard
	mask   int
}

type rwlockShard struct {
	_ [cacheLineSize]byte
	sync.RWMutex
}

// NewRWLock creates a new RWLock.
func NewRWLock() RWLock {
	shardsLen := runtime.GOMAXPROCS(0)
	shardsLen = int(internal.NextPowerOfTwo(uint(shardsLen)))
	mask := 1
	if shardsLen > 1 {
		mask = shardsLen - 1
	}
	lock := RWLock{
		shards: make([]rwlockShard, shardsLen),
		mask:   mask,
	}
	return lock
}

// Lock locks all underlying mutexes, preparing for exclusive access to
// the resource protected by the lock.
func (p RWLock) Lock() {
	for i := range p.shards {
		p.shards[i].Lock()
	}
}

// Unlock releases all underlying mutexes.
func (p RWLock) Unlock() {
	for i := range p.shards {
		p.shards[i].Unlock()
	}
}

// RLock acquires a non-exclusive reader lock, and returns the locker.
// The caller must hold the returned locker and calls it's Unlock method
// when it finishes work with the lock.
func (p RWLock) RLock() sync.Locker {
	pid := linkname.Runtime_procPin()
	linkname.Runtime_procUnpin()
	locker := p.shards[pid&p.mask].RLocker()
	locker.Lock()
	return locker
}
