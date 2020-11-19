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
	tmp := Int63()
	_ = Uint32()
	_ = Uint64()
	_ = Int31()
	_ = Int()
	_ = Int63n(tmp)
	_ = Int31n(int32(tmp))
	_ = Intn(int(tmp))
	_ = Float64()
	_ = Float32()
	_ = Perm(16)
	Shuffle(12, func(i, j int) {})
	_, err := Read(make([]byte, 16))
	assert.Nil(t, err)
	_ = NormFloat64()
	_ = ExpFloat64()
}
