package compress

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
)

func benchmarkInputs() map[string][]byte {
	random := make([]byte, 64<<10)
	_, _ = rand.New(rand.NewSource(1)).Read(random)
	json := bytes.Repeat([]byte(`{"id":123,"name":"example","enabled":true}`), 1024)
	encoded, _ := defaultGzipCodec.Compress(nil, json)
	return map[string][]byte{
		"short-text":         []byte("hello world"),
		"json":               json,
		"random":             random,
		"already-compressed": encoded,
	}
}

func BenchmarkCompressor(b *testing.B) {
	for name, input := range benchmarkInputs() {
		for _, threshold := range []int{0, DefaultThreshold} {
			b.Run(fmt.Sprintf("%s/threshold=%d", name, threshold), func(b *testing.B) {
				cfg := DefaultConfig()
				cfg.Threshold = threshold
				c := newTestCompressor(b, cfg)
				b.ReportAllocs()
				b.SetBytes(int64(len(input)))
				b.ResetTimer()
				var result CompressionResult
				for i := 0; i < b.N; i++ {
					result = c.Compress(context.Background(), input)
					if result.Err != nil {
						b.Fatal(result.Err)
					}
				}
				b.ReportMetric(float64(len(result.Data))/float64(len(input)), "frame-ratio")
			})
		}
	}
}

func BenchmarkCompressorParallel(b *testing.B) {
	c := newTestCompressor(b, DefaultConfig())
	input := benchmarkInputs()["json"]
	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result := c.Compress(context.Background(), input)
			if result.Err != nil {
				b.Error(result.Err)
				return
			}
			if _, _, err := c.Decompress(context.Background(), result.Data); err != nil {
				b.Error(err)
				return
			}
		}
	})
}

func BenchmarkDecompressor(b *testing.B) {
	c := newTestCompressor(b, DefaultConfig())
	for name, input := range benchmarkInputs() {
		frame := c.Compress(context.Background(), input).Data
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(input)))
			for i := 0; i < b.N; i++ {
				if _, _, err := c.Decompress(context.Background(), frame); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecompressorUnframed(b *testing.B) {
	c := newTestCompressor(b, CompressorConfig{})
	for name, input := range benchmarkInputs() {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(input)))
			for i := 0; i < b.N; i++ {
				if _, _, err := c.Decompress(context.Background(), input); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecompressorTypeNone(b *testing.B) {
	c := newTestCompressor(b, CompressorConfig{})
	for name, input := range benchmarkInputs() {
		frame := makeUncompressedFrame(input)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(input)))
			for i := 0; i < b.N; i++ {
				if _, _, err := c.Decompress(context.Background(), frame); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
