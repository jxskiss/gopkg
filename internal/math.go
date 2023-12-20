package internal

import (
	"time"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

// AddJitter adds random jitter to a duration.
//
// It adds or subtracts time from the duration within a given jitter fraction.
// For example for 10s and jitter 0.1, it returns a duration within [9s, 11s).
func AddJitter(duration time.Duration, jitter float64) time.Duration {
	// See math/rand/v2.Float64.
	randf64 := float64(linkname.Runtime_fastrand64()<<11>>11) / (1 << 53)
	x := jitter * (randf64*2 - 1)
	return time.Duration(float64(duration) * (1 + x))
}

func NextPowerOfTwo(x uint) uint {
	if x <= 1 {
		return 1
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32

	return x + 1
}
