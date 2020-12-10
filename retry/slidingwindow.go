package retry

import (
	"sync"
	"time"
)

const windowSize = 10
const rollingRetryThreshold = 30

var windowMap sync.Map

func getWindow(name string, overloadRatio float64) *window {
	w, ok := windowMap.Load(name)
	if !ok {
		w = &window{
			ratio: overloadRatio,
		}
		windowMap.Store(name, w)
	}
	return w.(*window)
}

type window struct {
	ratio float64
	succ  bucket
	fail  bucket
}

func (w *window) shouldRetry() bool {
	nowUnix := time.Now().Unix()
	count := w.fail.sum(nowUnix)
	if count > rollingRetryThreshold &&
		count > w.ratio*w.succ.sum(nowUnix) {
		return false
	}
	return true
}

type bucket struct {
	mu    sync.RWMutex
	index [windowSize]int64
	count [windowSize]float64
}

func (w *bucket) incr(nowUnix int64, i float64) {
	if i == 0 {
		return
	}
	idx := nowUnix % windowSize
	w.mu.Lock()
	if w.index[idx] != nowUnix {
		w.index[idx] = nowUnix
		w.count[idx] = 0
	}
	w.count[idx] += i
	w.mu.Unlock()
}

func (w *bucket) sum(nowUnix int64) float64 {
	var sum float64
	threshold := nowUnix - windowSize
	w.mu.RLock()
	for i := 0; i < windowSize; i++ {
		if w.index[i] >= threshold {
			sum += w.count[i]
		}
	}
	w.mu.RUnlock()
	return sum
}
