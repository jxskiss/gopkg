package retry

import (
	"github.com/jxskiss/gopkg/fastrand"
	"time"
)

// AddJitter adds random jitter to the duration.
//
// This adds or subtracts time from the duration within a given jitter fraction.
// For example for 10s and jitter 0.1, it will return a time within [9s, 11s])
func AddJitter(duration time.Duration, jitter float64) time.Duration {
	multiplier := jitter * (fastrand.Float64()*2 - 1)
	return time.Duration(float64(duration) * (1 + multiplier))
}

func Backoff(duration, max time.Duration, jitter float64) (double, withJitter time.Duration) {
	double = duration * 2
	if max > 0 && double > max {
		double = max
	}
	withJitter = AddJitter(double, jitter)
	return
}
