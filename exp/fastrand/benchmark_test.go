package fastrand

import (
	"math/rand"
	"testing"
)

func BenchmarkMathRand(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rand.Uint64()
	}
}

func BenchmarkRuntimeFastrand(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Fastrand()
	}
}

func BenchmarkPCG32(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Uint32()
	}
}

func BenchmarkPCG64(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Uint64()
	}
}
