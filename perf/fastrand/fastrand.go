// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fastrand

import (
	"github.com/jxskiss/gopkg/v2/internal/constraints"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
)

// runtimeSource is a Source that uses the runtime fastrand functions.
type runtimeSource struct{}

func (*runtimeSource) Uint64() uint64 {
	return linkname.Runtime_fastrand64()
}

// globalRand is the source of random numbers for the top-level
// convenience functions.
var globalRand = &Rand{src: &runtimeSource{}}

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func Uint64() (x uint64) { return globalRand.Uint64() }

// Int64 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int64() int64 { return globalRand.Int64() }

// Uint32 returns a pseudo-random 32-bit value as a uint32.
func Uint32() uint32 { return globalRand.Uint32() }

// Int32 returns a non-negative pseudo-random 31-bit integer as an int32.
func Int32() int32 { return globalRand.Int32() }

// Int returns a non-negative pseudo-random int from the default Source.
func Int() int { return globalRand.Int() }

// Float64 returns, as a float64, a pseudo-random number in the half-open interval [0.0,1.0).
func Float64() float64 { return globalRand.Float64() }

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers
// in the half-open interval [0,n).
func Perm(n int) []int { return globalRand.Perm(n) }

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) { globalRand.Shuffle(n, swap) }

// N returns a pseudo-random number in the half-open interval [0,n) from the default Source.
// The type parameter Int can be any integer type.
// It panics if n <= 0.
func N[Int constraints.Integer](n Int) Int {
	if n <= 0 {
		panic("fastrand: invalid argument to N")
	}
	return Int(uint64n(globalRand, uint64(n)))
}
