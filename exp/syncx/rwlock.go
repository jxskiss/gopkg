package syncx

import (
	"github.com/jxskiss/gopkg/internal/linkname"
	"runtime"
	"sync"
)

const cacheLineSize = 64

var shardsLen int

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

// RLocker returns a reading locker, preparing for sharing access to
// the resource protected by the lock.
//
// The caller must hold the returned locker and calls it's Unlock method
// when it don't need the lock.
func (p RWLock) RLocker() sync.Locker {
	pid := linkname.Runtime_procPin()
	linkname.Runtime_procUnpin()
	return p[pid%shardsLen].RLocker()
}
