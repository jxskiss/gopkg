package rthash

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
)

func TestHash(t *testing.T) {
	cases := []any{
		"abc",
		int8(19),
		uint8(19),
		int16(12345),
		uint16(12345),
		int32(8484848),
		uint32(8484848),
		int64(1234567890),
		uint64(1234567890),
		int(1234567890),
		uint64(1234567890),
		uintptr(1234567890),
		float32(1.1314),
		float64(1.1314),
		complex(float32(1.1314), float32(1.1314)),
		complex(float64(1.1314), float64(1.1314)),
		hashable{1234, "1234"},
	}

	var h = New()
	for _, x := range cases {
		hash := h.Hash(x)
		log.Println(fmt.Sprintf("%T: %v, hash: %d", x, x, hash))
	}
}

type hashable struct {
	A int
	B string
}

func BenchmarkHash_Int64(b *testing.B) {
	h := New()
	x := rand.Int63()
	for i := 0; i < b.N; i++ {
		_ = h.Int64(x)
	}
}

func BenchmarkHash_Bytes(b *testing.B) {
	h := New()
	x := []byte("this is a short sample text")
	for i := 0; i < b.N; i++ {
		_ = h.Bytes(x)
	}
}
