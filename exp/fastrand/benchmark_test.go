package fastrand

import (
	"math/rand"
	"testing"
)

func BenchmarkMathRandUint32(b *testing.B) {
	s := rand.NewSource(int64(Uint64()))
	g := rand.New(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint32()
	}
}

func BenchmarkMathRandUint64(b *testing.B) {
	s := rand.NewSource(int64(Uint64()))
	g := rand.New(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint64()
	}
}

func BenchmarkPCG32(b *testing.B) {
	g := NewPCG32()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint32()
	}
}

func BenchmarkPCG64(b *testing.B) {
	g := NewPCG64()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint64()
	}
}

// ---------------------- bounded ---------------------- //

const boundN = 123456

func BenchmarkMathRand_Uint32n(b *testing.B) {
	s := rand.NewSource(int64(Uint64()))
	g := rand.New(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Int31n(boundN)
	}
}

func BenchmarkMathRand_Uint64n(b *testing.B) {
	s := rand.NewSource(int64(Uint64()))
	g := rand.New(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Int63n(boundN)
	}
}

func BenchmarkPCG32_Uint32n(b *testing.B) {
	g := NewPCG32()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint32n(boundN)
	}
}

func BenchmarkPCG32_Uint32nRough(b *testing.B) {
	g := NewPCG32()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint32nRough(boundN)
	}
}

func BenchmarkPCG64_Uint64n(b *testing.B) {
	g := NewPCG64()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint64n(boundN)
	}
}

func BenchmarkPCG64_Uint64nRough(b *testing.B) {
	g := NewPCG64()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Uint64nRough(boundN)
	}
}

// ----------------- concurrently safe ----------------- //

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
