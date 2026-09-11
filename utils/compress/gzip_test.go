package compress

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"math"
	"testing"
)

func TestGzipCodecLevels(t *testing.T) {
	for level := gzip.HuffmanOnly; level <= gzip.BestCompression; level++ {
		codec, err := NewGzipCodec(level)
		if err != nil || codec.CompressionLevel() != level {
			t.Fatalf("level %d: %v", level, err)
		}
		for _, input := range [][]byte{nil, []byte("hello"), []byte("你好世界"), bytes.Repeat([]byte("a"), 10000)} {
			dst := make([]byte, 3, 1024)
			copy(dst, "pre")
			encoded, err := codec.Compress(dst, input)
			if err != nil || string(encoded[:3]) != "pre" {
				t.Fatalf("compress: %v", err)
			}
			out, err := codec.Decompress(encoded[3:], max(1, len(input)))
			if err != nil || !bytes.Equal(out, input) {
				t.Fatalf("round trip level %d: %v", level, err)
			}
			reader, err := gzip.NewReader(bytes.NewReader(encoded[3:]))
			if err != nil {
				t.Fatal(err)
			}
			standard, err := io.ReadAll(reader)
			reader.Close()
			if err != nil || !bytes.Equal(standard, input) {
				t.Fatalf("standard decode: %v", err)
			}
		}
	}
	for _, level := range []int{-3, 10, math.MaxInt} {
		if codec, err := NewGzipCodec(level); codec != nil || !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("level %d: %v", level, err)
		}
	}
}

func TestGzipCodecCorruptionAndReuse(t *testing.T) {
	codec, _ := NewGzipCodec(gzip.DefaultCompression)
	input := bytes.Repeat([]byte("a"), 1024)
	valid, _ := codec.Compress(nil, input)
	invalid := [][]byte{nil, {}, []byte("invalid")}
	for i := 0; i < len(valid); i++ {
		invalid = append(invalid, bytes.Clone(valid[:i]))
	}
	crc := bytes.Clone(valid)
	crc[len(crc)-8] ^= 1
	size := bytes.Clone(valid)
	size[len(size)-4] ^= 1
	invalid = append(invalid, crc, size, append(bytes.Clone(valid), 1))
	for i, data := range invalid {
		if out, err := codec.Decompress(data, len(input)); err == nil || out != nil {
			t.Fatalf("accepted corrupt stream %d", i)
		}
		out, err := codec.Decompress(valid, len(input))
		if err != nil || !bytes.Equal(out, input) {
			t.Fatalf("reuse after corruption %d: %v", i, err)
		}
	}
	if _, err := codec.Decompress(crc, len(input)); !errors.Is(err, gzip.ErrChecksum) {
		t.Fatalf("checksum error lost: %v", err)
	}
}

func TestGzipCodecLimits(t *testing.T) {
	codec, _ := NewGzipCodec(gzip.BestSpeed)
	input := bytes.Repeat([]byte("a"), 4096)
	valid, _ := codec.Compress(nil, input)
	for _, limit := range []int{1, 4095, 4096, 4097, math.MaxInt} {
		out, err := codec.Decompress(valid, limit)
		if limit < len(input) {
			if out != nil || !errors.Is(err, ErrDecodedTooLarge) {
				t.Fatalf("limit %d: %v", limit, err)
			}
		} else if err != nil || !bytes.Equal(out, input) {
			t.Fatalf("limit %d: %v", limit, err)
		}
	}
	for _, limit := range []int{-1, 0} {
		if out, err := codec.Decompress(valid, limit); out != nil || !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("limit %d: %v", limit, err)
		}
	}
	multi := append(bytes.Clone(valid), valid...)
	if out, err := codec.Decompress(multi, len(input)); out != nil || !errors.Is(err, ErrDecodedTooLarge) {
		t.Fatalf("multistream limit: %v", err)
	}
	out, err := codec.Decompress(multi, len(input)*2)
	if err != nil || !bytes.Equal(out, bytes.Repeat(input, 2)) {
		t.Fatalf("multistream: %v", err)
	}
}

func FuzzGzipCodecDecompress(f *testing.F) {
	codec, _ := NewGzipCodec(gzip.DefaultCompression)
	valid, _ := codec.Compress(nil, []byte("hello"))
	f.Add(valid)
	f.Add([]byte{})
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
