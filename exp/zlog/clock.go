package zlog

import (
	"sync/atomic"
	"time"
	"unsafe"

	"go.uber.org/zap/zapcore"
)

func newMilliClock(milli int) zapcore.Clock {
	now := time.Now()
	clock := &milliClock{
		now: unsafe.Pointer(&now),
	}
	go clock.tick(time.Duration(milli) * time.Millisecond)
	return clock
}

type milliClock struct {
	now unsafe.Pointer
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

func (m *milliClock) NewTicker(duration time.Duration) *time.Ticker {
	return time.NewTicker(duration)
}
