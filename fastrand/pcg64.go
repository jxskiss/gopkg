package fastrand

import (
	"github.com/jxskiss/gopkg/internal/linkname"
	"math/bits"
)

// PCG64 is an implementation of a 64-bit permuted congruential
// generator as defined in
//
// 	PCG: A Family of Simple Fast Space-Efficient Statistically Good
// 	Algorithms for Random Number Generation
// 	Melissa E. O’Neill, Harvey Mudd College
// 	http://www.pcg-random.org/pdf/toms-oneill-pcg-family-v1.02.pdf
//
// The generator here is the congruential generator PCG XSL RR 128/64 (LCG)
// as found in the software available at http://www.pcg-random.org/.
// It has period 2^128 with 128 bits of state, producing 64-bit values.
// It's state is represented by two uint64 words.
//
// The implementation is forked from
// https://github.com/golang/exp/blob/master/rand/rng.go
type PCG64 struct {
	low  uint64
	high uint64
}

const (
	maxUint64 = (1 << 64) - 1
	int63Mask = (1 << 63) - 1
	int31Mask = (1 << 31) - 1

	multiplier = 47026247687942121848144207491837523525
	mulHigh    = multiplier >> 64
	mulLow     = multiplier & maxUint64

	increment = 117397592171526113268558934119004209487
	incHigh   = increment >> 64
	incLow    = increment & maxUint64

	initializer = 245720598905631564143578724636268694099
	initHigh    = initializer >> 64
	initLow     = initializer & maxUint64
)

// NewPCG64 returns a PCG64 generator initialized with random state
// and sequence.
func NewPCG64() *PCG64 {
	a, b := linkname.Runtime_fastrand(), linkname.Runtime_fastrand()
	c, d := linkname.Runtime_fastrand(), linkname.Runtime_fastrand()
	low := uint64(a)<<32 + uint64(b)
	high := uint64(c)<<32 + uint64(d)
	return &PCG64{low: low, high: high}
}

// Seed uses the provided seed value to initialize the generator to a deterministic state.
func (p *PCG64) Seed(low, high uint64) {
	p.low = low
	p.high = high
}

func (p *PCG64) add() {
	var carry uint64
	p.low, carry = bits.Add64(p.low, incLow, 0)
	p.high, _ = bits.Add64(p.high, incHigh, carry)
}

func (p *PCG64) multiply() {
	hi, lo := bits.Mul64(p.low, mulLow)
	hi += p.high * mulLow
	hi += p.low * mulHigh
	p.low = lo
	p.high = hi
}

// Uint64 returns a pseudo-random 64-bit unsigned integer as a uint64.
func (p *PCG64) Uint64() uint64 {
	p.multiply()
	p.add()
	// XOR high and low 64 bits together and rotate right by high 6 bits of state.
	return bits.RotateLeft64(p.high^p.low, -int(p.high>>58))
}

// Uint32 returns a pseudo-random 32-bit unsigned integer as a uint32.
func (p *PCG64) Uint32() uint32 {
	return uint32(p.Uint64() >> 32)
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (p *PCG64) Int63() int64 {
	return int64(p.Uint64() & int63Mask)
}

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func (p *PCG64) Int31() int32 {
	return int32(p.Int63() >> 32)
}

// Int returns a non-negative pseudo-random int.
func (p *PCG64) Int() int {
	u := uint(p.Int63())
	return int(u << 1 >> 1) // clear sign bit if int == int32
}

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
//
// For implementation details, see:
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction
// https://lemire.me/blog/2016/06/30/fast-random-shuffling
func (p *PCG64) Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int63n")
	}

	u64 := p.Uint64()
	hi, lo := bits.Mul64(u64, uint64(n))
	if lo < uint64(n) {
		threshold := uint64(-n) % uint64(n)
		for lo < threshold {
			u64 = p.Uint64()
			hi, lo = bits.Mul64(u64, uint64(n))
		}
	}
	return int64(hi)
}

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
//
// For implementation details, see:
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction
// https://lemire.me/blog/2016/06/30/fast-random-shuffling
func (p *PCG64) Int31n(n int32) int32 {
	if n <= 0 {
		panic("invalid argument to Int32n")
	}

	u32 := p.Uint32()
	prod := uint64(u32) * uint64(n)
	low := uint32(prod)
	if low < uint32(n) {
		thresh := uint32(-n) % uint32(n)
		for low < thresh {
			u32 = p.Uint32()
			prod = uint64(u32) * uint64(n)
			low = uint32(prod)
		}
	}
	return int32(prod >> 32)
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (p *PCG64) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	if n <= 1<<31-1 {
		return int(p.Int31n(int32(n)))
	}
	return int(p.Int63n(int64(n)))
}

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func (p *PCG64) Float64() float64 {
	return float64(p.Int63n(1<<53)) / (1 << 53)
}

// Float32 returns, as a float32, a pseudo-random number in [0.0,1.0).
func (p *PCG64) Float32() float32 {
	return float32(p.Int31n(1<<24)) / (1 << 24)
}

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers [0,n).
func (p *PCG64) Perm(n int) []int {
	m := make([]int, n)
	for i := 1; i < n; i++ {
		j := p.Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func (p *PCG64) Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(p.Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(p.Int31n(int32(i + 1)))
		swap(i, j)
	}
}
