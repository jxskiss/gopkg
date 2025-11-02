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

func init() {
	compress.ProvideDecompressor(defaultZstdAlg)
}

var defaultZstdAlg = NewZstdCompressor(DefaultCompression)

type ZstdCompressor struct {
	level int
}

// NewZstdCompressor creates a new ZstdCompressor instance.
// level specifies the compression level, valid values are [1, 20].
func NewZstdCompressor(level int) compress.CompressionAlg {
	if level == 0 {
		level = DefaultCompression
	} else if level < BestSpeed {
		level = BestSpeed
	} else if level > BestCompression {
		level = BestCompression
	}
	return &ZstdCompressor{
		level: level,
	}
}

func (p *ZstdCompressor) Type() compress.AlgType {
	return compress.TypeZstd
}

func (p *ZstdCompressor) CompressionLevel() int {
	return p.level
}

// Compress compresses the source byte slice using the zstd algorithm.
func (p *ZstdCompressor) Compress(dst []byte, data []byte) ([]byte, error) {
	bound := zstd.CompressBound(len(data))
	out := make([]byte, len(dst), len(dst)+bound)
	copy(out, dst)
	tmp, err := zstd.CompressLevel(out[len(dst):], data, p.level)
	if err != nil {
		return nil, fmt.Errorf("zstd compress failed: %w", err)
	}
	out = out[:len(dst)+len(tmp)]
	return out, nil
}

// Decompress decompresses the source byte slice using the zstd algorithm.
func (p *ZstdCompressor) Decompress(src []byte) ([]byte, error) {
	out, err := zstd.Decompress(nil, src)
	if err != nil {
		return nil, fmt.Errorf("zstd decompress failed: %w", err)
	}
	return out, nil
}
