package internal

import "github.com/jxskiss/gopkg/v2/internal/linkname"

// Functions in this file are copies from perf/fastrand to use internally,
// see perf/fastrand for detail docs.

func Uint64() (x uint64) {
	return linkname.Runtime_fastrand64()
}

func Int63() (x int64) {
	return int64(Uint64() & (1<<63 - 1))
}

func Int63n(n int64) (x int64) {
	if n <= 0 {
		panic("bug (internal): Int63n got invalid argument n <= 0")
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return Int63() & (n - 1)
	}
	_max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := Int63()
	for v > _max {
		v = Int63()
	}
	return v % n
}

func Float64() (x float64) {
	return float64(Int63n(1<<53)) / (1 << 53)
}
