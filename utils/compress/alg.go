package compress

import (
	"errors"
	"fmt"
)

type AlgType byte

const (
	TypeUnknown    AlgType = '?'
	TypeNoCompress AlgType = '0'
	TypeGzip       AlgType = '1'
	TypeZstd       AlgType = '2'
)

func (alg AlgType) String() string {
	switch alg {
	case TypeUnknown:
		return "unknown"
	case TypeNoCompress:
		return "no_compress"
	case TypeGzip:
		return "gzip"
	case TypeZstd:
		return "zstd"
	default:
		return fmt.Sprintf("alg_type_%d", alg)
	}
}

const (
	ReasonEmptyData      = "emptyData"
	ReasonBelowThreshold = "belowThreshold"
	ReasonSavingNegative = "savingNegative"
	ReasonSavingTooSmall = "savingTooSmall"
	ReasonCompressFailed = "compressFailed"
)

var (
	headerByte2      = byte('\x01')
	headerNoCompress = []byte{byte(TypeNoCompress), headerByte2}
	gzipMagicNumber  = []byte("\x1f\x8b\x08")
	zstdMagicNumber  = []byte("\x28\xb5\x2f\xfd")
)

var (
	errZstdDecompressorNotAvailable = errors.New("zstd decompressor not available, check if zstd package is imported")
)

// CompressionAlg is the interface for compression algorithm.
// A CompressionAlg implementation must be safe for concurrent use by multiple goroutines.
type CompressionAlg interface {
	Type() AlgType
	CompressionLevel() int
	Compress(dst []byte, data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
}

var (
	gzipDecompressFunc func([]byte) ([]byte, error)
	zstdDecompressFunc func([]byte) ([]byte, error)
)

func ProvideDecompressor(alg CompressionAlg) {
	switch alg.Type() {
	case TypeGzip:
		gzipDecompressFunc = alg.Decompress
	case TypeZstd:
		zstdDecompressFunc = alg.Decompress
	default:
		panic(fmt.Sprintf("unsupported compression type %s", alg.Type()))
	}
}

func decompressZstd(data []byte) (decompressed []byte, compressType AlgType, err error) {
	if zstdDecompressFunc == nil {
		return nil, TypeZstd, errZstdDecompressorNotAvailable
	}
	decompressed, err = zstdDecompressFunc(data)
	return decompressed, TypeZstd, err
}
