package compress

import (
	"errors"
	"fmt"
)

type AlgType byte

const (
	TypeUnknown AlgType = 0
	TypeNone    AlgType = '0'
	TypeGzip    AlgType = '1'
	TypeZstd    AlgType = '2'
)

func (alg AlgType) String() string {
	switch alg {
	case TypeUnknown:
		return "unknown"
	case TypeNone:
		return "none"
	case TypeGzip:
		return "gzip"
	case TypeZstd:
		return "zstd"
	default:
		return fmt.Sprintf("alg_type_%d", alg)
	}
}

func isSupportedAlg(alg AlgType) bool {
	return alg == TypeGzip || alg == TypeZstd
}

var (
	ErrInvalidFrame         = errors.New("invalid compression frame")
	ErrUnsupportedAlgorithm = errors.New("unsupported compression algorithm")
	ErrDecoderUnavailable   = errors.New("compression decoder unavailable")
	ErrDecodedTooLarge      = errors.New("decoded data exceeds size limit")
	ErrInvalidConfig        = errors.New("invalid compression configuration")
)

// Decoder implementations must be safe for concurrent use and keep Type stable.
type Decoder interface {
	Type() AlgType
	// Decompress decodes a raw algorithm stream, with a positive output limit.
	// It must enforce the limit during decoding, return ErrDecodedTooLarge on
	// overflow, validate the complete stream, and return nil data on error.
	// Successful output must not alias data or internal reusable storage.
	Decompress(data []byte, maxDecodedSize int) ([]byte, error)
}

// Codec implementations must be safe for concurrent use.
// A codec's decoder must accept its encoder's output.
// Only TypeGzip and TypeZstd are supported currently.
type Codec interface {
	Decoder
	// Compress appends a raw stream to dst, preserving its existing prefix.
	// data must not overlap dst and must not be modified.
	// Output may alias dst, but must remain valid across subsequent calls.
	// On error dst may be changed.
	Compress(dst, data []byte) ([]byte, error)
}
