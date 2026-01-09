package lptime

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/functicker"
)

const (
	defaultPrecision = time.Second
	minPrecision     = 10 * time.Millisecond
)

var (
	nowNano int64

	mu    sync.Mutex
	state struct {
		precision time.Duration
		ticker    *functicker.Ticker
	}
)

func init() {
	state.precision = defaultPrecision
	state.ticker = functicker.New(defaultPrecision, func() {
		atomic.StoreInt64(&nowNano, time.Now().UnixNano())
	})
	atomic.StoreInt64(&nowNano, time.Now().UnixNano())
}

func SetPrecision(precision time.Duration) {
	if precision <= 0 {
		panic("non-positive precision for lptime.SetPrecision")
	}
	mu.Lock()
	defer mu.Unlock()

	precision = max(precision, minPrecision)
	if precision < state.precision {
		state.precision = precision
		state.ticker.Reset(precision)
	}
}

func Now() time.Time {
	now := atomic.LoadInt64(&nowNano)
	return time.Unix(0, now)
}

func Unix() int64 {
	return atomic.LoadInt64(&nowNano) / int64(time.Second)
}

func UnixMilli() int64 {
	return atomic.LoadInt64(&nowNano) / int64(time.Millisecond)
}

func UnixMicro() int64 {
	return atomic.LoadInt64(&nowNano) / int64(time.Microsecond)
}

func UnixNano() int64 {
	return atomic.LoadInt64(&nowNano)
}
