package wheel

import "time"

type Ticker struct {
	C <-chan time.Time
	r *timer
}

func (t *Ticker) Stop() {
	t.r.w.delTimer(t.r)
}

func (t *Ticker) Reset(d time.Duration) {
	t.r.w.resetTimer(t.r, d, d)
}

func NewTicker(d time.Duration) *Ticker {
	return defaultWheel().NewTicker(d)
}

func Tick(d time.Duration) <-chan time.Time {
	return defaultWheel().Tick(d)
}

func TickFunc(d time.Duration, f func()) *Ticker {
	return defaultWheel().TickFunc(d, f)
}
