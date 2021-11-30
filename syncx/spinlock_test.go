package syncx

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

type naiveSpinLock uint32

func (sl *naiveSpinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		runtime.Gosched()
	}
}

func (sl *naiveSpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func newNaiveSpinLock() sync.Locker {
	return new(naiveSpinLock)
}

func BenchmarkSyncMutex(b *testing.B) {
	m := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			//nolint:staticcheck
			m.Unlock()
		}
	})
}

func BenchmarkNaiveSpinLock(b *testing.B) {
	spin := newNaiveSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			//nolint:staticcheck
			spin.Unlock()
		}
	})
}

func BenchmarkBackoffSpinLock(b *testing.B) {
	spin := NewSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			//nolint:staticcheck
			spin.Unlock()
		}
	})
}
