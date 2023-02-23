package retry

import (
	"sync"
	"time"
)

const windowSize = 10
const rollingRetryThreshold = 30

var breakerMap sync.Map

type breakerKey struct {
	name          string
	overloadRatio float64
}

func getBreaker(name string, overloadRatio float64) *breaker {
	key := breakerKey{
		name:          name,
		overloadRatio: overloadRatio,
	}
	w, ok := breakerMap.Load(key)
	if !ok {
		w = &breaker{
			ratio: overloadRatio,
		}
		breakerMap.Store(key, w)
	}
	return w.(*breaker)
}

type breaker struct {
	ratio float64
	succ  bucket
	fail  bucket
}

func (b *breaker) shouldRetry() bool {
	nowUnix := time.Now().Unix()
	count := b.fail.sum(nowUnix)
	if count > rollingRetryThreshold &&
		count > b.ratio*b.succ.sum(nowUnix) {
		return false
	}
	return true
}

type bucket struct {
	mu    sync.RWMutex
	index [windowSize]int64
	count [windowSize]float64
}

func (b *bucket) incr(nowUnix int64) {
	idx := nowUnix % windowSize
	b.mu.Lock()
	if b.index[idx] != nowUnix {
		b.index[idx] = nowUnix
		b.count[idx] = 0
	}
	b.count[idx]++
	b.mu.Unlock()
}

func (b *bucket) sum(nowUnix int64) float64 {
	var sum float64
	threshold := nowUnix - windowSize
	b.mu.RLock()
	for i := 0; i < windowSize; i++ {
		if b.index[i] >= threshold {
			sum += b.count[i]
		}
	}
	b.mu.RUnlock()
	return sum
}
