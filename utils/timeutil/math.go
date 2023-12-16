package timeutil

import (
	"time"

	"github.com/jxskiss/gopkg/v2/internal"
)

// AddJitter adds random jitter to a duration.
//
// It adds or subtracts time from the duration within a given jitter fraction.
// For example for 10s and jitter 0.1, it returns a duration within [9s, 11s).
func AddJitter(duration time.Duration, jitter float64) time.Duration {
	return internal.AddJitter(duration, jitter)
}

// Backoff doubles the given duration. If max_ is given larger than 0 and
// the doubled value is greater than max_, it will be limited to max_.
// The param jitter can be used to add random jitter to the doubled duration.
func Backoff(duration, max time.Duration, jitter float64) (double, withJitter time.Duration) {
	double = duration * 2
	if max > 0 && double > max {
		double = max
	}
	withJitter = AddJitter(double, jitter)
	return
}
