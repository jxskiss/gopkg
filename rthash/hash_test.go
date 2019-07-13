package rthash

import (
	"fmt"
	"log"
	"testing"
)

func Test_Hash(t *testing.T) {
	cases := []interface{}{
		"abc",
		[]byte("abc"),
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

	var h uintptr
	for _, x := range cases {
		h = Hash(x)
		log.Println(fmt.Sprintf("%T: %v, hash: %d", x, x, h))
	}
}

type hashable struct {
	A int
	B string
}
