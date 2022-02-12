package clock

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Default is the default clock which uses the system clock for all
// operations.
var Default = systemClock{}

// Clock is a source of time and ticker.
type Clock interface {

	// Now returns the current local time.
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

type milliClock struct {
	now unsafe.Pointer // *time.Time
}

var (
	milliClockMu  sync.Mutex
	milliClockMap = make(map[int]*milliClock)
)

// NewMilliClock returns a Clock instance which ticks at milliseconds
// precision but gives better performance.
//
// Generally, the system clock is preferred in most cases.
func NewMilliClock(milli int) *milliClock {
	milliClockMu.Lock()
	defer milliClockMu.Unlock()
	clock := milliClockMap[milli]
	if clock == nil {
		now := time.Now()
		clock = &milliClock{
			now: unsafe.Pointer(&now),
		}
		duration := time.Duration(milli) * time.Millisecond
		go clock.tick(duration)
		milliClockMap[milli] = clock
	}
	return clock
}

func (m *milliClock) tick(d time.Duration) {
	ticker := time.NewTicker(d)
	for t := range ticker.C {
		tCopy := t
		atomic.StorePointer(&m.now, unsafe.Pointer(&tCopy))
	}
}

func (m *milliClock) Now() time.Time {
	return *(*time.Time)(atomic.LoadPointer(&m.now))
}
