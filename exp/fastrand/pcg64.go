package fastrand

import "math/bits"

// pcg64 is an implementation of a 64-bit permuted congruential
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
// Is state is represented by two uint64 words.
//
// https://github.com/golang/exp/blob/master/rand/rng.go
type pcg64 struct {
	low  uint64
	high uint64
}

const (
	maxUint64 = (1 << 64) - 1

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

func newPCG64() *pcg64 {
	return &pcg64{low: initLow, high: initHigh}
}

// Seed uses the provided seed value to initialize the generator to a deterministic state.
func (p *pcg64) Seed(low, high uint64) {
	p.low = low
	p.high = high
}

// Uint64 returns a pseudo-random 64-bit unsigned integer as a uint64.
func (p *pcg64) Uint64() uint64 {
	p.multiply()
	p.add()
	// XOR high and low 64 bits together and rotate right by high 6 bits of state.
	return bits.RotateLeft64(p.high^p.low, -int(p.high>>58))
}

func (p *pcg64) add() {
	var carry uint64
	p.low, carry = bits.Add64(p.low, incLow, 0)
	p.high, _ = bits.Add64(p.high, incHigh, carry)
}

func (p *pcg64) multiply() {
	hi, lo := bits.Mul64(p.low, mulLow)
	hi += p.high * mulLow
	hi += p.low * mulHigh
	p.low = lo
	p.high = hi
}
