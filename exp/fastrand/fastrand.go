package fastrand

import "runtime"

const cacheLineSize = 64

var (
	shardsLen   int
	globalPCG32 pinPCG32
	globalPCG64 pinPCG64
)

func init() {
	// Multiply by 4 to try best to cover the case that user may change
	// the maximum CPU numbers.
	shardsLen = 4 * runtime.GOMAXPROCS(0)

	globalPCG32 = make(pinPCG32, shardsLen)
	for i := 0; i < len(globalPCG32); i++ {
		a, b, c, d := runtime_fastrand(), runtime_fastrand(), runtime_fastrand(), runtime_fastrand()
		state := uint64(a)<<32 + uint64(b)
		seq := uint64(c)<<32 + uint64(d)
		globalPCG32[i].Seed(state, seq)
	}
	globalPCG64 = make(pinPCG64, shardsLen)
	for i := 0; i < len(globalPCG64); i++ {
		a, b, c, d := runtime_fastrand(), runtime_fastrand(), runtime_fastrand(), runtime_fastrand()
		low := uint64(a)<<32 + uint64(b)
		high := uint64(c)<<32 + uint64(d)
		globalPCG64[i].Seed(low, high)
	}
}

type pcg32Source struct {
	_ [cacheLineSize]byte
	pcg32
}

type pinPCG32 []pcg32Source

type pcg64Source struct {
	_ [cacheLineSize]byte
	pcg64
}

type pinPCG64 []pcg64Source

// Uint32 returns a pseudo-random 32-bit value as a uint32
// from the default pcg32 source.
func Uint32() (x uint32) {
	pid := runtime_procPin()
	x = globalPCG32[pid].Uint32()
	runtime_procUnpin()
	return
}

// Uint32n returns a pseudo-random unsigned 32-bit integer in range [0, n)
// from the default pcg32 source.
// It panics if n <= 0.
func Uint32n(n uint32) (x uint32) {
	pid := runtime_procPin()
	x = globalPCG32[pid].Uint32n(n)
	runtime_procUnpin()
	return
}

// Uint64 returns a pseudo-random 64-bit value as a uint64
// from the default pcg64 source.
func Uint64() (x uint64) {
	pid := runtime_procPin()
	x = globalPCG64[pid].Uint64()
	runtime_procUnpin()
	return
}

// Uint64n returns a pseudo-random unsigned 64-bit integer in range [0, n)
// from the default pcg64 source.
// It panics if n <= 0.
func Uint64n(n uint64) (x uint64) {
	pid := runtime_procPin()
	x = globalPCG64[pid].Uint64n(n)
	runtime_procUnpin()
	return
}
