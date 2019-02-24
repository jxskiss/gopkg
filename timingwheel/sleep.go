package wheel

import "time"

type Timer struct {
	C <-chan time.Time
	r *timer
}

func (t *Timer) Stop() {
	t.r.w.delTimer(t.r)
}

func (t *Timer) Reset(d time.Duration) {
	t.r.w.resetTimer(t.r, d, 0)
}

func NewTimer(d time.Duration) *Timer {
	return defaultWheel().NewTimer(d)
}

func Sleep(d time.Duration) {
	defaultWheel().Sleep(d)
}

func After(d time.Duration) <-chan time.Time {
	return defaultWheel().After(d)
}

func AfterFunc(d time.Duration, f func()) *Timer {
	return defaultWheel().AfterFunc(d, f)
}
