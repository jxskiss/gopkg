package functicker

import "time"

type Ticker struct {
	t *time.Ticker
}

func New(d time.Duration, f func()) *Ticker {
	timer := time.AfterFunc(time.Hour, f)
	timer.Stop()
	ticker := (*time.Ticker)(timer)
	ticker.Reset(d)
	return &Ticker{t: ticker}
}

func (t *Ticker) Stop() {
	t.t.Stop()
}

func (t *Ticker) Reset(d time.Duration) {
	t.t.Reset(d)
}
