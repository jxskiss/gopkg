package fastrand

import "github.com/jxskiss/gopkg/v2/internal/linkname"

// Uint64 returns a pseudo-random 64-bit unsigned integer as a uint64.
func Uint64() (x uint64) {
	return linkname.Runtime_fastrand64()
}

// Uint32 returns a pseudo-random 32-bit unsigned integer as a uint32.
func Uint32() (x uint32) {
	return linkname.Runtime_fastrand()
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int63() (x int64) {
	return int64(Uint64() & (1<<63 - 1))
}

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func Int31() (x int32) {
	return int32(Uint32() & (1<<31 - 1))
}

// Int returns a non-negative pseudo-random int.
func Int() (x int) {
	u := uint(Int63())
	return int(u << 1 >> 1) // clear sign bit if int == int32
}

// Int63n returns, as an int64, a non-negative pseudo-random number in the half-open interval [0,n).
// It panics if n <= 0.
func Int63n(n int64) (x int64) {
	if n <= 0 {
		panic("fastrand: invalid argument to Int63n")
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

// Int31n returns, as an int32, a non-negative pseudo-random number in the half-open interval [0,n).
// It panics if n <= 0.
//
// For implementation details, see:
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction
// https://lemire.me/blog/2016/06/30/fast-random-shuffling
func Int31n(n int32) (x int32) {
	if n <= 0 {
		panic("fastrand: invalid argument to Int31n")
	}

	u32 := Uint32()
	prod := uint64(u32) * uint64(n)
	low := uint32(prod)
	if low < uint32(n) {
		threshold := uint32(-n) % uint32(n)
		for low < threshold {
			u32 = Uint32()
			prod = uint64(u32) * uint64(n)
			low = uint32(prod)
		}
	}
	return int32(prod >> 32)
}

// Intn returns, as an int, a non-negative pseudo-random number in the half-open interval [0,n).
// It panics if n <= 0.
func Intn(n int) (x int) {
	if n <= 0 {
		panic("fastrand: invalid argument to Intn")
	}
	if n <= 1<<31-1 {
		return int(Int31n(int32(n)))
	}
	return int(Int63n(int64(n)))
}

// Float64 returns, as a float64, a pseudo-random number in the half-open interval [0.0,1.0).
func Float64() (x float64) {
	return float64(Int63n(1<<53)) / (1 << 53)
}

// Float32 returns, as a float32, a pseudo-random number in the half-open interval [0.0,1.0).
func Float32() (x float32) {
	return float32(Int31n(1<<24)) / (1 << 24)
}

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers
// in the half-open interval [0,n).
func Perm(n int) []int {
	m := make([]int, n)
	for i := 1; i < n; i++ {
		j := Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// Param swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("fastrand: invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(Int31n(int32(i + 1)))
		swap(i, j)
	}
}

// Read generates len(p) random bytes and writes them into p.
// It always returns len(p) and a nil error.
//
// Read is safe for concurrent use.
func Read(p []byte) (n int, err error) {
	if len(p) > 0 {
		_wyread(Uint32(), p)
	}
	return len(p), nil
}
