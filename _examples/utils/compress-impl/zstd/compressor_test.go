package zstd

import (
	"bytes"
	"context"
	"errors"
	"math"
	"sync"
	"testing"

	"github.com/jxskiss/gopkg/v2/utils/compress"
)

func TestZstdCodec(t *testing.T) {
	for _, level := range []int{BestSpeed, DefaultCompression, BestCompression} {
		codec, err := NewZstdCodec(level)
		if err != nil {
			t.Fatal(err)
		}
		if codec.CompressionLevel() != level {
			t.Fatal("level changed")
		}
		for _, input := range [][]byte{nil, []byte("hello"), []byte("你好，世界"), bytes.Repeat([]byte("zstd data"), 1000)} {
			prefix := make([]byte, 3, 128)
			copy(prefix, "pre")
			encoded, err := codec.Compress(prefix, input)
			if err != nil || len(encoded) < 3 || string(encoded[:3]) != "pre" {
				t.Fatalf("compress: %v", err)
			}
			decoded, err := codec.Decompress(encoded[3:], max(1, len(input)))
			if err != nil || !bytes.Equal(decoded, input) {
				t.Fatalf("round trip: %v", err)
			}
		}
	}
	for _, level := range []int{0, -1, BestCompression + 1} {
		if codec, err := NewZstdCodec(level); codec != nil || !errors.Is(err, compress.ErrInvalidConfig) {
			t.Fatalf("level %d: %v", level, err)
		}
	}
}

func TestZstdCodecLimitsAndCorruption(t *testing.T) {
	codec, _ := NewZstdCodec(DefaultCompression)
	input := bytes.Repeat([]byte("zstd data"), 1000)
	valid, _ := codec.Compress(nil, input)
	for _, limit := range []int{1, len(input) - 1, len(input), len(input) + 1, math.MaxInt} {
		out, err := codec.Decompress(valid, limit)
		if limit < len(input) {
			if out != nil || !errors.Is(err, compress.ErrDecodedTooLarge) {
				t.Fatalf("limit %d: %v", limit, err)
			}
		} else if err != nil || !bytes.Equal(out, input) {
			t.Fatalf("limit %d: %v", limit, err)
		}
	}
	for _, limit := range []int{-1, 0} {
		if out, err := codec.Decompress(valid, limit); out != nil || !errors.Is(err, compress.ErrInvalidConfig) {
			t.Fatalf("limit %d: %v", limit, err)
		}
	}
	for i := 0; i < len(valid); i++ {
		if out, err := codec.Decompress(valid[:i], len(input)); err == nil || out != nil {
			t.Fatalf("accepted truncation %d", i)
		}
	}
	for _, data := range [][]byte{[]byte("invalid"), append(bytes.Clone(valid), 1)} {
		if out, err := codec.Decompress(data, len(input)); err == nil || out != nil {
			t.Fatal("accepted invalid stream")
		}
	}
	multi := append(bytes.Clone(valid), valid...)
	if out, err := codec.Decompress(multi, len(input)); out != nil || !errors.Is(err, compress.ErrDecodedTooLarge) {
		t.Fatalf("multistream limit: %v", err)
	}
	if out, err := codec.Decompress(multi, 2*len(input)); err != nil || !bytes.Equal(out, bytes.Repeat(input, 2)) {
		t.Fatalf("multistream: %v", err)
	}
}

func TestZstdCompressorConcurrency(t *testing.T) {
	codec, _ := NewZstdCodec(BestSpeed)
	cfg := compress.DefaultConfig()
	cfg.Codec = codec
	c, err := compress.NewCompressor(cfg)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := compress.NewCompressor(compress.CompressorConfig{Decoders: []compress.Decoder{codec}})
	if err != nil {
		t.Fatal(err)
	}
	input := bytes.Repeat([]byte("concurrent zstd"), 1000)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				result := c.Compress(context.Background(), input)
				if result.Err != nil || result.Type != compress.TypeZstd || result.Reason != "" {
					t.Errorf("compression: %s %s %v", result.Type, result.Reason, result.Err)
					return
				}
				for _, decoder := range []compress.Compressor{c, reader} {
					out, typ, err := decoder.Decompress(context.Background(), result.Data)
					if err != nil || typ != compress.TypeZstd || !bytes.Equal(out, input) {
						t.Errorf("round trip: %s %v", typ, err)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

func TestZstdCodecOwnership(t *testing.T) {
	codec, _ := NewZstdCodec(BestSpeed)
	input := bytes.Repeat([]byte("a"), 1000)
	expected := bytes.Clone(input)
	encoded, err := codec.Compress(nil, input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 'b'
	out, err := codec.Decompress(encoded, len(expected))
	if err != nil || !bytes.Equal(out, expected) {
		t.Fatalf("input alias: %v", err)
	}
	out[0] = 'c'
	again, err := codec.Decompress(encoded, len(expected))
	if err != nil || !bytes.Equal(again, expected) {
		t.Fatalf("output alias: %v", err)
	}
	saved := bytes.Clone(encoded)
	if _, err := codec.Compress(nil, input); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, saved) {
		t.Fatal("output changed after reuse")
	}
}

func FuzzZstdCodecRoundTrip(f *testing.F) {
	codec, _ := NewZstdCodec(BestSpeed)
	f.Add([]byte("hello world"))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 1<<20 {
			t.Skip()
		}
		encoded, err := codec.Compress(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		out, err := codec.Decompress(encoded, max(1, len(input)))
		if err != nil || !bytes.Equal(out, input) {
			t.Fatalf("round trip: %v", err)
		}
	})
}

func FuzzZstdCodecDecompress(f *testing.F) {
	codec, _ := NewZstdCodec(BestSpeed)
	f.Add([]byte{})
	valid, _ := codec.Compress(nil, []byte("hello"))
	f.Add(valid)
	f.Fuzz(func(t *testing.T, data []byte) {
		out, err := codec.Decompress(data, 64<<10)
		if err != nil && out != nil {
			t.Fatal("partial output on error")
		}
		if len(out) > 64<<10 {
			t.Fatal("limit exceeded")
		}
	})
}
