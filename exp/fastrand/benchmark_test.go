package fastrand

import (
	"math/rand"
	"testing"
)

func BenchmarkConcurrentRuntimeFastrand(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Fastrand()
		}
	})
}

func BenchmarkConcurrentMathRandUint32(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rand.Uint32()
		}
	})
}

func BenchmarkConcurrentMathRandUint64(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rand.Uint64()
		}
	})
}

func BenchmarkConcurrentPCG32(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Uint32()
		}
	})
}

func BenchmarkConcurrentPCG64(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Uint64()
		}
	})
}
