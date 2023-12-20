// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fastrand

import "math/bits"

// https://numpy.org/devdocs/reference/random/upgrading-pcg64.html
// https://github.com/imneme/pcg-cpp/commit/871d0494ee9c9a7b7c43f753e3d8ca47c26f8005

// NewPCG returns a new PCG generator seeded with the given values.
func NewPCG(seed1, seed2 uint64) *Rand {
	return &Rand{
		src: &pcgSource{seed1, seed2},
	}
}

// A pcgSource is a PCG generator with 128 bits of internal state.
// A zero pcgSource is equivalent to pcgSource{0, 0}.
type pcgSource struct {
	hi uint64
	lo uint64
}

func (p *pcgSource) next() (hi, lo uint64) {
	// https://github.com/imneme/pcg-cpp/blob/428802d1a5/include/pcg_random.hpp#L161
	//
	// Numpy's pcgSource multiplies by the 64-bit value cheapMul
	// instead of the 128-bit value used here and in the official pcgSource code.
	// This does not seem worthwhile, at least for Go: not having any high
	// bits in the multiplier reduces the effect of low bits on the highest bits,
	// and it only saves 1 multiply out of 3.
	// (On 32-bit systems, it saves 1 out of 6, since Mul64 is doing 4.)
	const (
		mulHi = 2549297995355413924
		mulLo = 4865540595714422341
		incHi = 6364136223846793005
		incLo = 1442695040888963407
	)

	// state = state * mul + inc
	hi, lo = bits.Mul64(p.lo, mulLo)
	hi += p.hi*mulLo + p.lo*mulHi
	lo, c := bits.Add64(lo, incLo, 0)
	hi, _ = bits.Add64(hi, incHi, c)
	p.lo = lo
	p.hi = hi
	return hi, lo
}

// Uint64 return a uniformly-distributed random uint64 value.
func (p *pcgSource) Uint64() uint64 {
	hi, lo := p.next()

	// XSL-RR would be
	//	hi, lo := p.next()
	//	return bits.RotateLeft64(lo^hi, -int(hi>>58))
	// but Numpy uses DXSM and O'Neill suggests doing the same.
	// See https://github.com/golang/go/issues/21835#issuecomment-739065688
	// and following comments.

	// DXSM "double xorshift multiply"
	// https://github.com/imneme/pcg-cpp/blob/428802d1a5/include/pcg_random.hpp#L1015

	// https://github.com/imneme/pcg-cpp/blob/428802d1a5/include/pcg_random.hpp#L176
	const cheapMul = 0xda942042e4dd58b5
	hi ^= hi >> 32
	hi *= cheapMul
	hi ^= hi >> 48
	hi *= (lo | 1)
	return hi
}
