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
	// From Go 1.23, time.Ticker and time.Timer are defined as different structs,
	// though the underlying data is still same, we have to use unsafe trick
	// to cast the timer pointer.
	//
	// !!! FOR FURTHER NEW VERSIONS, MUST CHECK THIS !!!

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
