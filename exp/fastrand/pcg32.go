package fastrand

// pcg32 is an implementation of a 32-bit permuted congruential generator.
//
// Developed by Melissa O'Neill <oneill@pcg-random.org>
// Paper and details at http://www.pcg-random.org
// Ported to Go by Michael Jones <michael.jones@gmail.com>
//
// https://github.com/MichaelTJones/pcg
type pcg32 struct {
	state     uint64
	increment uint64
}

const (
	pcg32State      = 0x853c49e6748fea9b //  9600629759793949339
	pcg32Increment  = 0xda3e39cb94b95bdb // 15726070495360670683
	pcg32Multiplier = 0x5851f42d4c957f2d //  6364136223846793005
)

func newPCG32() *pcg32 {
	return &pcg32{pcg32State, pcg32Increment}
}

// Seed uses the provided seed value to initialize the generator to a deterministic state.
func (p *pcg32) Seed(state, sequence uint64) *pcg32 {
	p.increment = (sequence << 1) | 1
	p.state = (state+p.increment)*pcg32Multiplier + p.increment
	return p
}

// Uint32 returns a pseudo-random 32-bit unsigned integer as a uint32.
func (p *pcg32) Uint32() uint32 {
	// Advance 64-bit linear congruential generator to new state.
	oldState := p.state
	p.state = oldState*pcg32Multiplier + p.increment

	// Confuse and permute 32-bit output from old state.
	xorShifted := uint32(((oldState >> 18) ^ oldState) >> 27)
	rot := uint32(oldState >> 59)

	return (xorShifted >> rot) | (xorShifted << ((-rot) & 31))
}
