package fastrand

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"runtime"
)

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

	// Use random bytes to seed the generators.
	randbuf := make([]byte, shardsLen*16*2)
	if _, err := rand.Read(randbuf); err != nil {
		panic(fmt.Sprintf("fastrand: rand.Read error: %v", err))
	}

	x := 0
	globalPCG32 = make(pinPCG32, shardsLen)
	for i := 0; i < len(globalPCG32); i++ {
		state := binary.BigEndian.Uint64(randbuf[x : x+8])
		seq := binary.BigEndian.Uint64(randbuf[x+8 : x+16])
		globalPCG32[i].Seed(state, seq)
		x += 16
	}
	globalPCG64 = make(pinPCG64, shardsLen)
	for i := 0; i < len(globalPCG64); i++ {
		low := binary.BigEndian.Uint64(randbuf[x : x+8])
		high := binary.BigEndian.Uint64(randbuf[x+8 : x+16])
		globalPCG64[i].Seed(low, high)
		x += 16
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
// from the default Source.
func Uint32() (x uint32) {
	pid := runtime_procPin()
	x = globalPCG32[pid].Uint32()
	runtime_procUnpin()
	return
}

// Uint64 returns a pseudo-random 64-bit value as a uint64
// from the default Source.
func Uint64() (x uint64) {
	pid := runtime_procPin()
	x = globalPCG64[pid].Uint64()
	runtime_procUnpin()
	return
}
