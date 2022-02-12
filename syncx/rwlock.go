package syncx

import (
	"runtime"
	"sync"

	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

const cacheLineSize = 64

var (
	shardsLen  int
	shardsMask int = 1
)

// RWLock holds a group of sharded RWMutex, it gives better performance
// in read-heavy workloads by reducing lock contention, but the performance
// for exclusive Lock is poor.
type RWLock []rwlockShard

type rwlockShard struct {
	_ [cacheLineSize]byte
	sync.RWMutex
}

func init() {
	shardsLen = runtime.GOMAXPROCS(0)
	shardsLen = int(internal.NextPowerOfTwo(uint(shardsLen)))
	if shardsLen > 1 {
		shardsMask = shardsLen - 1
	}
}

// NewRWLock creates a new RWLock.
func NewRWLock() RWLock {
	return make(RWLock, shardsLen)
}

// Lock locks all underlying mutexes, preparing for exclusive access to
// the resource protected by the lock.
func (p RWLock) Lock() {
	for i := range p {
		p[i].Lock()
	}
}

// Unlock releases all underlying mutexes.
func (p RWLock) Unlock() {
	for i := range p {
		p[i].Unlock()
	}
}

// RLock acquires a non-exclusive reader lock, and returns the locker.
// The caller must hold the returned locker and calls it's Unlock method
// when it finishes work with the lock.
func (p RWLock) RLock() sync.Locker {
	pid := linkname.Runtime_procPin()
	linkname.Runtime_procUnpin()
	locker := p[pid&shardsMask].RLocker()
	locker.Lock()
	return locker
}
