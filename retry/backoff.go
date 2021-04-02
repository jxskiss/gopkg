package retry

import (
	"math/rand"
	"time"
)

// AddJitter adds random jitter to the duration.
//
// This adds or subtracts time from the duration within a given jitter fraction.
// For example for 10s and jitter 0.1, it will return a time within [9s, 11s])
func AddJitter(duration time.Duration, jitter float64) time.Duration {
	multiplier := jitter * (rand.Float64()*2 - 1)
	return time.Duration(float64(duration) * (1 + multiplier))
}

// Backoff doubles the given duration. If max is given larger than 0, and
// the doubled value is greater than max, it will be limited to max. The
// param jitter can be used to add random jitter to the doubled duration.
func Backoff(duration, max time.Duration, jitter float64) (double, withJitter time.Duration) {
	double = duration * 2
	if max > 0 && double > max {
		double = max
	}
	withJitter = AddJitter(double, jitter)
	return
}
