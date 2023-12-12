// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fastrand

import (
	"math/bits"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func Uint64() (x uint64) {
	return globalSource.Uint64()
}

// Int64 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int64() int64 { return int64(Uint64() &^ (1 << 63)) }

// Float64 returns, as a float64, a pseudo-random number in the half-open interval [0.0,1.0).
func Float64() float64 {
	// There are exactly 1<<53 float64s in [0,1). Use Intn(1<<53) / (1<<53).
	return float64(Uint64()<<11>>11) / (1 << 53)
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) {
	ShuffleWithSource(globalSource, n, swap)
}

func ShuffleWithSource(s Source, n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	for i := n - 1; i > 0; i-- {
		j := int(uint64n(s, uint64(i+1)))
		swap(i, j)
	}
}

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers
// in the half-open interval [0,n).
func Perm(n int) []int {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	Shuffle(len(p), func(i, j int) { p[i], p[j] = p[j], p[i] })
	return p
}

// N returns a pseudo-random number in the half-open interval [0,n) from the default Source.
// The type parameter Int can be any integer type.
// It panics if n <= 0.
func N[Int constraints.Integer](n Int) Int {
	if n <= 0 {
		panic("invalid argument to N")
	}
	return Int(uint64n(globalSource, uint64(n)))
}

const is32bit = ^uint(0)>>32 == 0

// uint64n is the no-bounds-checks version of Uint64N.
func uint64n(s Source, n uint64) uint64 {
	if is32bit && uint64(uint32(n)) == n {
		return uint64(uint32n(s, uint32(n)))
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return s.Uint64() & (n - 1)
	}

	// Suppose we have a uint64 x uniform in the range [0,2⁶⁴)
	// and want to reduce it to the range [0,n) preserving exact uniformity.
	// We can simulate a scaling arbitrary precision x * (n/2⁶⁴) by
	// the high bits of a double-width multiply of x*n, meaning (x*n)/2⁶⁴.
	// Since there are 2⁶⁴ possible inputs x and only n possible outputs,
	// the output is necessarily biased if n does not divide 2⁶⁴.
	// In general (x*n)/2⁶⁴ = k for x*n in [k*2⁶⁴,(k+1)*2⁶⁴).
	// There are either floor(2⁶⁴/n) or ceil(2⁶⁴/n) possible products
	// in that range, depending on k.
	// But suppose we reject the sample and try again when
	// x*n is in [k*2⁶⁴, k*2⁶⁴+(2⁶⁴%n)), meaning rejecting fewer than n possible
	// outcomes out of the 2⁶⁴.
	// Now there are exactly floor(2⁶⁴/n) possible ways to produce
	// each output value k, so we've restored uniformity.
	// To get valid uint64 math, 2⁶⁴ % n = (2⁶⁴ - n) % n = -n % n,
	// so the direct implementation of this algorithm would be:
	//
	//	hi, lo := bits.Mul64(r.Uint64(), n)
	//	thresh := -n % n
	//	for lo < thresh {
	//		hi, lo = bits.Mul64(r.Uint64(), n)
	//	}
	//
	// That still leaves an expensive 64-bit division that we would rather avoid.
	// We know that thresh < n, and n is usually much less than 2⁶⁴, so we can
	// avoid the last four lines unless lo < n.
	//
	// See also:
	// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction
	// https://lemire.me/blog/2016/06/30/fast-random-shuffling
	hi, lo := bits.Mul64(s.Uint64(), n)
	if lo < n {
		thresh := -n % n
		for lo < thresh {
			hi, lo = bits.Mul64(s.Uint64(), n)
		}
	}
	return hi
}

// uint32n is an identical computation to uint64n
// but optimized for 32-bit systems.
func uint32n(s Source, n uint32) uint32 {
	if n&(n-1) == 0 { // n is power of two, can mask
		return uint32(s.Uint64()) & (n - 1)
	}

	// On 64-bit systems we still use the uint64 code below because
	// the probability of a random uint64 lo being < a uint32 n is near zero,
	// meaning the unbiasing loop almost never runs.
	// On 32-bit systems, here we need to implement that same logic in 32-bit math,
	// both to preserve the exact output sequence observed on 64-bit machines
	// and to preserve the optimization that the unbiasing loop almost never runs.
	//
	// We want to compute
	// 	hi, lo := bits.Mul64(r.Uint64(), n)
	// In terms of 32-bit halves, this is:
	// 	x1:x0 := r.Uint64()
	// 	0:hi, lo1:lo0 := bits.Mul64(x1:x0, 0:n)
	// Writing out the multiplication in terms of bits.Mul32 allows
	// using direct hardware instructions and avoiding
	// the computations involving these zeros.
	x := s.Uint64()
	lo1a, lo0 := bits.Mul32(uint32(x), n)
	hi, lo1b := bits.Mul32(uint32(x>>32), n)
	lo1, c := bits.Add32(lo1a, lo1b, 0)
	hi += c
	if lo1 == 0 && lo0 < uint32(n) {
		n64 := uint64(n)
		thresh := uint32(-n64 % n64)
		for lo1 == 0 && lo0 < thresh {
			x := s.Uint64()
			lo1a, lo0 = bits.Mul32(uint32(x), n)
			hi, lo1b = bits.Mul32(uint32(x>>32), n)
			lo1, c = bits.Add32(lo1a, lo1b, 0)
			hi += c
		}
	}
	return hi
}

// Read generates len(b) random bytes and writes them into b.
// It always returns len(b) and a nil error.
func Read(b []byte) (n int, err error) {
	if len(b) > 0 {
		wyhashRead(Uint64(), b)
	}
	return len(b), nil
}

// wyrand: https://github.com/wangyi-fudan/wyhash
type wyrand uint64

//nolint:unused
const (
	wyp0 = 0xa0761d6478bd642f
	wyp1 = 0xe7037ed1a0b428db
	wyp2 = 0x8ebc6af09c88c6e3
	wyp3 = 0x589965cc75374cc3
	wyp4 = 0x1d8e4e27c47d124f
)

func (r *wyrand) Uint64() uint64 {
	*r += wyp0
	hi, lo := bits.Mul64(uint64(*r), uint64(*r)^wyp1)
	return hi ^ lo
}

func wyhashRead(seed uint64, p []byte) {
	r := wyrand(seed)
	intp := *(*[]uint64)(unsafe.Pointer(&p))
	var i, end int
	for i, end = 0, len(p)-8; i < end; i += 8 {
		intp[i>>3] = r.Uint64()
	}
	if i < len(p) {
		u64 := r.Uint64()
		for j := range p[i:] {
			p[i+j] = byte(u64 >> (j * 8))
		}
	}
}
