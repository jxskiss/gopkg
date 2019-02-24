package retry

import "time"

type strategy func(time.Duration) time.Duration

func exp(sleep time.Duration) time.Duration {
	return sleep * 2
}

func constant(sleep time.Duration) time.Duration {
	return sleep
}

type linear struct {
	step time.Duration
}

func (l linear) next(sleep time.Duration) time.Duration {
	return sleep + l.step
}
