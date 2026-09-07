package zstd

import (
	"fmt"

	"github.com/DataDog/zstd"

	"github.com/jxskiss/gopkg/v2/utils/compress"
)

const (
	BestSpeed          = zstd.BestSpeed
	BestCompression    = zstd.BestCompression
	DefaultCompression = zstd.DefaultCompression
)

// NewZstdCodec accepts levels in range [BestSpeed, BestCompression].
func NewZstdCodec(level int) (*ZstdCodec, error) {
	if level < BestSpeed || level > BestCompression {
		return nil, fmt.Errorf("%w: zstd level %d", compress.ErrInvalidConfig, level)
	}
	return &ZstdCodec{level: level}, nil
}

type ZstdCodec struct{ level int }

func (p *ZstdCodec) Type() compress.AlgType { return compress.TypeZstd }
func (p *ZstdCodec) CompressionLevel() int  { return p.level }

func (p *ZstdCodec) Compress(dst, data []byte) ([]byte, error) {
	bound := zstd.CompressBound(len(data))
	out := make([]byte, len(dst), len(dst)+bound)
	copy(out, dst)
	encoded, err := zstd.CompressLevel(out[len(dst):], data, p.level)
	if err != nil {
		return nil, fmt.Errorf("zstd compress: %w", err)
	}
	return out[:len(dst)+len(encoded)], nil
}

func (p *ZstdCodec) Decompress(data []byte, maxDecodedSize int) ([]byte, error) {
	if maxDecodedSize <= 0 {
		return nil, fmt.Errorf("%w: max decoded size must be positive", compress.ErrInvalidConfig)
	}
	if len(data) == 0 {
		return nil, zstd.ErrEmptySlice
	}
	// DataDog's streaming reader can accept truncated frames. zstd.DecompressInto
	// validates whole frames and never allocates an unbounded output buffer.
	size := min(1024, maxDecodedSize)
	for {
		out := make([]byte, size)
		n, err := zstd.DecompressInto(out, data)
		if err == nil {
			return out[:n], nil
		}
		if !zstd.IsDstSizeTooSmallError(err) {
			return nil, fmt.Errorf("zstd decompress: %w", err)
		}
		if size == maxDecodedSize {
			return nil, compress.ErrDecodedTooLarge
		}
		size += min(size, maxDecodedSize-size)
	}
}
