package retry

import (
	"math/rand"
	"sync"
	"time"
)

var (
	// pseudoRand is safe for concurrent use.
	pseudoRand *lockedMathRand
)

func init() {
	pseudoRand = &lockedMathRand{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// AddJitter adds random jitter to the duration.
//
// This adds or subtracts time from the duration within a given jitter fraction.
// For example for 10s and jitter 0.1, it will return a time within [9s, 11s])
func AddJitter(duration time.Duration, jitter float64) time.Duration {
	multiplier := jitter * (pseudoRand.Float64()*2 - 1)
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

type lockedMathRand struct {
	sync.Mutex
	rnd *rand.Rand
}

func (r *lockedMathRand) Int63n(max int64) int64 {
	r.Lock()
	n := r.rnd.Int63n(max)
	r.Unlock()
	return n
}

func (r *lockedMathRand) Float64() float64 {
	r.Lock()
	n := r.rnd.Float64()
	r.Unlock()
	return n
}
