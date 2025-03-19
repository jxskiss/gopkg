package functicker

import (
	"time"
	"unsafe"
)

type Ticker struct {
	t *time.Ticker
}

func New(d time.Duration, f func()) *Ticker {
	ticker := newTimeTicker(f)
	ticker.Reset(d)
	return &Ticker{t: ticker}
}

func newTimeTicker(f func()) *time.Ticker {
	timer := time.AfterFunc(time.Hour, f)
	timer.Stop()
	ticker := (*time.Ticker)(unsafe.Pointer(timer))
	return ticker
}

func (t *Ticker) Stop() {
	t.t.Stop()
}

func (t *Ticker) Reset(d time.Duration) {
	t.t.Reset(d)
}
