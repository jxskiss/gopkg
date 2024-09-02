package rthash

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestHashFunc(t *testing.T) {
	testCases := []struct {
		value    any
		hashFunc any
	}{
		{"abc", NewHashFunc[string]()},
		{int8(19), NewHashFunc[int8]()},
		{uint8(19), NewHashFunc[uint8]()},
		{int16(12345), NewHashFunc[int16]()},
		{uint16(12345), NewHashFunc[uint16]()},
		{int32(8484848), NewHashFunc[int32]()},
		{uint32(8484848), NewHashFunc[uint32]()},
		{int64(1234567890), NewHashFunc[int64]()},
		{uint64(1234567890), NewHashFunc[uint64]()},
		{int(1234567890), NewHashFunc[int]()},
		{uint64(1234567890), NewHashFunc[uint64]()},
		{uintptr(1234567890), NewHashFunc[uintptr]()},
		{float32(1.1314), NewHashFunc[float32]()},
		{float64(1.1314), NewHashFunc[float64]()},
		{complex(float32(1.1314), float32(1.1314)), NewHashFunc[complex64]()},
		{complex(float64(1.1314), float64(1.1314)), NewHashFunc[complex128]()},
		{hashable{1234, "1234"}, NewHashFunc[hashable]()},
	}

	for _, tc := range testCases {
		hash := reflect.ValueOf(tc.hashFunc).
			Call([]reflect.Value{reflect.ValueOf(tc.value)})[0].Interface().(uintptr)
		t.Logf("%T: %v, hash: %d", tc.value, tc.value, hash)
	}
}

type hashable struct {
	A int
	B string
}

func BenchmarkHashFunc_Int64(b *testing.B) {
	f := NewHashFunc[int64]()
	x := rand.Int63()
	for i := 0; i < b.N; i++ {
		_ = f(x)
	}
}

func BenchmarkHashFunc_String(b *testing.B) {
	f := NewHashFunc[string]()
	x := "this is a short sample text"
	for i := 0; i < b.N; i++ {
		_ = f(x)
	}
}
