package fastrand

import (
	"math/rand"
	"testing"
)

func BenchmarkRuntimeFastrand_Uint32(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Uint32()
	}
}

func BenchmarkMathRand_Uint32(b *testing.B) {
	s := rand.NewSource(int64(Uint64()))
	g := rand.New(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint32()
	}
}

func BenchmarkPCG64_Uint32(b *testing.B) {
	g := NewPCG64()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint32()
	}
}

// ---------------------- bounded ---------------------- //

const boundN = 123456

func BenchmarkRuntimeFastrand_Int63n(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Int63n(boundN)
	}
}

func BenchmarkMathRand_Int63n(b *testing.B) {
	s := rand.NewSource(int64(Uint64()))
	g := rand.New(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Int63n(boundN)
	}
}

func BenchmarkPCG64_Int63n(b *testing.B) {
	g := NewPCG64()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Int63n(boundN)
	}
}

// ----------------- concurrent safe ----------------- //

func BenchmarkConcurrentRuntimeFastrand_Uint32(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Uint32()
		}
	})
}

func BenchmarkConcurrentRuntimeFastrand_Uint64(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Uint64()
		}
	})
}

func BenchmarkConcurrentMathRand_Uint32(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rand.Uint32()
		}
	})
}

func BenchmarkConcurrentMathRand_Uint64(b *testing.B) {
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rand.Uint64()
		}
	})
}
