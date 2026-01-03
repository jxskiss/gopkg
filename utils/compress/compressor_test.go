package compress

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressor(t *testing.T) {
	ctx := context.Background()
	data := []byte("hello world, this is a test string for compression.")

	t.Run("DefaultCompressor", func(t *testing.T) {
		compressor := DefaultCompressor
		testData := bytes.Repeat(data, 1024)
		compressedData, compressType, _ := compressor.Compress(ctx, testData)
		assert.NotEqual(t, testData, compressedData)
		assert.Equal(t, TypeGzip, compressType)

		decompressedData, _, err := compressor.Decompress(ctx, compressedData)
		assert.Nil(t, err)
		assert.Equal(t, testData, decompressedData)
	})

	t.Run("custom compressor - below threshold", func(t *testing.T) {
		compressor := NewCompressor(CompressorConfig{
			Threshold: 1024,
		})
		compressedData, compressType, reason := compressor.Compress(ctx, data)
		assert.Equal(t, TypeNoCompress, compressType)
		assert.Equal(t, ReasonBelowThreshold, reason)
		decompressedData, _, err := compressor.Decompress(ctx, compressedData)
		assert.Nil(t, err)
		assert.Equal(t, data, decompressedData)
	})

	t.Run("custom compressor - small saving", func(t *testing.T) {
		compressor := NewCompressor(CompressorConfig{
			MinSaving: 0.9,
			Threshold: 1,
		})
		testData := bytes.Repeat(data, 2)
		compressedData, compressType, reason := compressor.Compress(ctx, testData)
		assert.Equal(t, TypeNoCompress, compressType)
		assert.Equal(t, ReasonSavingTooSmall, reason)
		decompressedData, _, err := compressor.Decompress(ctx, compressedData)
		assert.Nil(t, err)
		assert.Equal(t, testData, decompressedData)
	})

	t.Run("empty data", func(t *testing.T) {
		compressor := DefaultCompressor
		compressedData, compressType, reason := compressor.Compress(ctx, []byte{})
		assert.Equal(t, TypeNoCompress, compressType)
		assert.Equal(t, ReasonEmptyData, reason)
		assert.Empty(t, compressedData)
		decompressedData, _, err := compressor.Decompress(ctx, compressedData)
		assert.Nil(t, err)
		assert.Empty(t, decompressedData)
	})

	t.Run("decompress no header", func(t *testing.T) {
		compressor := NewCompressor(CompressorConfig{
			TreatNoHeaderAsNoCompress: true,
		})
		decompressedData, compressType, err := compressor.Decompress(ctx, data)
		assert.Nil(t, err)
		assert.Equal(t, TypeNoCompress, compressType)
		assert.Equal(t, data, decompressedData)
	})

	t.Run("decompress unknown header", func(t *testing.T) {
		compressor := DefaultCompressor
		badData := []byte{0x05, 0x06, 0x07, 0x08}
		_, _, err := compressor.Decompress(ctx, badData)
		assert.NotNil(t, err)
	})
}
