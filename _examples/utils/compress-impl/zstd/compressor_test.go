package zstd

import (
	"bytes"
	"context"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/utils/compress"
)

func TestZstdCompressor(t *testing.T) {
	ctx := context.Background()
	zstdImpl := NewZstdCompressor(BestCompression)
	compressor := compress.NewCompressor(compress.CompressorConfig{
		BizName:   "test",
		Alg:       zstdImpl,
		Threshold: 1,
	})
	original := make([]byte, 11*1024+333)
	_, err := rand.Read(original[:1024])
	if err != nil {
		t.Fatalf("Read random data error: %v", err)
	}
	compressed, compressType, failReason := compressor.Compress(ctx, original)
	if failReason != "" {
		t.Fatalf("Compress error: %v", failReason)
	}
	if compressType != compress.TypeZstd {
		t.Fatalf("Compress type is not Zstd")
	}
	decompressed, compressType, err := compressor.Decompress(ctx, compressed)
	if err != nil {
		t.Fatalf("Decompress error: %v", err)
	}
	if compressType != compress.TypeZstd {
		t.Fatalf("Decompress type is not Zstd")
	}
	if !bytes.Equal(original, decompressed) {
		t.Fatalf("Decompressed data does not match original data")
	}
}

// Test concurrent use of the compressor to ensure pool safety
func TestZstdCompressor_Concurrency(t *testing.T) {
	compressor := compress.NewCompressor(compress.CompressorConfig{
		BizName:   "test",
		Alg:       defaultZstdAlg,
		Threshold: 1,
		MinSaving: 0.1,
	})
	originalData := []byte(strings.Repeat("concurrency test data", 500))

	numGoroutines := 100
	results := make(chan []byte, numGoroutines)
	failReasons := make(chan string, numGoroutines)
	errs := make(chan error, numGoroutines)

	ctx := context.Background()
	for i := 0; i < numGoroutines; i++ {
		go func() {
			compressed, _, reason := compressor.Compress(ctx, originalData)
			if reason != "" {
				failReasons <- reason
				return
			}
			decompressed, _, err := compressor.Decompress(ctx, compressed)
			if err != nil {
				errs <- err
				return
			}
			results <- decompressed
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		select {
		case reason := <-failReasons:
			if reason == compress.ReasonCompressFailed {
				t.Errorf("Concurrency test failed with reason: %s", reason)
			}
		case err := <-errs:
			t.Errorf("Concurrency test failed with error: %v", err)
		case decompressed := <-results:
			assert.Equal(t, originalData, decompressed, "Decompressed data should match original in concurrent test")
		}
	}
}

func FuzzZstdCompressor(f *testing.F) {
	compressor := NewZstdCompressor(BestSpeed)
	f.Add([]byte("hello world"))
	f.Fuzz(func(t *testing.T, data []byte) {
		compressed, err := compressor.Compress(nil, data)
		if err != nil {
			t.Fatalf("Compress error: %v", err)
		}
		decompressed, err := compressor.Decompress(compressed)
		if err != nil {
			t.Fatalf("Decompress error: %v", err)
		}
		if !bytes.Equal(data, decompressed) {
			t.Fatalf("Decompressed data does not match original data")
		}
	})
}
