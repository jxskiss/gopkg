package compress

import (
	"compress/gzip"
	"context"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipCompressor_CompressDecompress(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "small string",
			data: []byte("hello world"),
		},
		{
			name: "long compressible string",
			data: []byte(strings.Repeat("a", 1000)),
		},
		{
			name: "chinese characters",
			data: []byte("你好，世界！这是一段中文测试文本。"),
		},
		{
			name: "medium compressible data",
			data: []byte(strings.Repeat("Go is a statically typed, compiled programming language designed at Google", 50)),
		},
		{
			name: "large compressible data",
			data: []byte(strings.Repeat("compressed data test case that should be very compressible", 1000)),
		},
	}

	ctx := context.Background()
	algImpl := NewGzipCompressor(gzip.DefaultCompression)
	compressor := NewCompressor(CompressorConfig{
		Alg:       algImpl,
		BizName:   "test",
		Threshold: 1,
		MinSaving: 0.0001,
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressedData, compressType, failReason := compressor.Compress(ctx, tt.data)
			if compressType == TypeNoCompress {
				assert.NotEmpty(t, failReason)
				assert.Equal(t, TypeNoCompress, AlgType(compressedData[0]))
				assert.Equal(t, tt.data, compressedData[2:])
			} else {
				assert.Empty(t, failReason, "Compress should not fail")
				assert.Equal(t, TypeGzip, compressType, "Compress type should be TypeGzip")
				assert.NotNil(t, compressedData, "Compressed data should not be nil")
				assert.Equal(t, TypeGzip, AlgType(compressedData[0]), "Compressed data should be TypeGzip")
			}

			if len(tt.data) > 0 && len(compressedData) >= len(tt.data) && len(tt.data) > 20 { // gzip header is about 10-20 bytes
				t.Logf("Original size: %d, Compressed size: %d", len(tt.data), len(compressedData))
			}

			decompressedData, _, err := compressor.Decompress(ctx, compressedData)
			assert.NoError(t, err, "Decompress should not return an error")
			assert.Equal(t, tt.data, decompressedData, "Decompressed data should match original data")
		})
	}

	// Test with incompressible data (random bytes)
	t.Run("random data (incompressible)", func(t *testing.T) {
		data := make([]byte, 1000)
		rand.Read(data)

		compressedData, compressType, failReason := compressor.Compress(ctx, data)
		assert.NotEmpty(t, failReason, "Compress should fail")
		assert.Equal(t, TypeNoCompress, compressType, "Compress type should be TypeNoCompress")
		assert.NotNil(t, compressedData)
		assert.Equal(t, TypeNoCompress, AlgType(compressedData[0]))
		assert.Equal(t, data, compressedData[2:])

		// Incompressible data might result in larger compressed size due to gzip overhead
		// assert.True(t, len(compressedData) > len(data), "Compressed random data should be larger")
		t.Logf("Original random size: %d, Compressed random size: %d", len(data), len(compressedData))

		decompressedData, compressType, err := compressor.Decompress(ctx, compressedData)
		assert.NoError(t, err)
		assert.Equal(t, TypeNoCompress, compressType)
		assert.Equal(t, data, decompressedData)
	})
}

func TestGzipCompressor_DecompressError(t *testing.T) {
	compressor := NewGzipCompressor(gzip.DefaultCompression)

	t.Run("invalid gzip data", func(t *testing.T) {
		invalidData := []byte("this is not gzip data")
		_, err := compressor.Decompress(invalidData)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "gzip.NewReader failed") ||
			strings.Contains(err.Error(), "reset gzip.Reader failed"))
	})

	t.Run("empty compressed data", func(t *testing.T) {
		emptyData := []byte{1, 2, 3, 4, 5} // Not enough for a valid gzip header
		_, err := compressor.Decompress(emptyData)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "gzip.NewReader failed") ||
			strings.Contains(err.Error(), "reset gzip.Reader failed"))
	})
}

// Test concurrent use of the compressor to ensure pool safety
func TestGzipCompressor_Concurrency(t *testing.T) {
	compressor := NewCompressor(CompressorConfig{
		BizName:   "test",
		Alg:       defaultGzipAlg,
		Threshold: 0,
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
			if reason == ReasonCompressFailed {
				t.Errorf("Concurrency test failed with reason: %s", reason)
			}
		case err := <-errs:
			t.Errorf("Concurrency test failed with error: %v", err)
		case decompressed := <-results:
			assert.Equal(t, originalData, decompressed, "Decompressed data should match original in concurrent test")
		}
	}
}
