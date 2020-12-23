package fastrand

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRand(t *testing.T) {
	rnd := New(time.Now().UnixNano())
	_ = rnd.Int63()
}

func TestRandFunctions(t *testing.T) {
	Seed(time.Now().UnixNano())
	var tmp64 int64
	var tmp32 int32
	for tmp64 <= 0 || tmp32 <= 0 {
		tmp64 = Int63()
		tmp32 = int32(tmp64)
	}
	_ = Uint32()
	_ = Uint64()
	_ = Int31()
	_ = Int()
	_ = Int63n(tmp64)
	_ = Int31n(tmp32)
	_ = Intn(int(tmp32))
	_ = Float64()
	_ = Float32()
	_ = Perm(16)
	Shuffle(12, func(i, j int) {})
	_, err := Read(make([]byte, 16))
	assert.Nil(t, err)
	_ = NormFloat64()
	_ = ExpFloat64()
}
