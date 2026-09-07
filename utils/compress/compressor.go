package compress

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

const (
	DefaultMinReductionRatio = 0.05
	DefaultThreshold         = 5 * 1024
	DefaultMaxDecodedSize    = 64 * 1024 * 1024

	frameMagic           = "GPKC"
	frameVersion         = byte(1)
	frameVersionOffset   = len(frameMagic)
	frameAlgorithmOffset = frameVersionOffset + 1
	frameSizeOffset      = frameAlgorithmOffset + 1
	frameHeaderSize      = frameSizeOffset + 8
)

// DefaultCompressor uses gzip with DefaultConfig. Replacing it does not change
// constructor defaults; replacement must happen before concurrent access.
var DefaultCompressor = newDefaultCompressor()

func newDefaultCompressor() Compressor {
	p, err := NewCompressor(DefaultConfig())
	if err != nil {
		panic(err)
	}
	return p
}

// Compressor encodes and decodes versioned frames.
// All successful outputs are owned by the caller and do not alias input.
// Methods are safe for concurrent use if configured codecs and callbacks are.
// Context is only passed to the callback; these in-memory operations
// do not support cancellation.
type Compressor interface {
	Compress(ctx context.Context, data []byte) CompressionResult
	// Decompress accepts only complete frames, never bare streams or plaintext.
	// It returns nil data on error; TypeUnknown means the algorithm was not parsed.
	Decompress(ctx context.Context, data []byte) ([]byte, AlgType, error)
}

type CompressorConfig struct {
	// BizName optionally identifies the business scenario in CompressionStats
	// passed to CompressCallback, allowing metrics and logs to distinguish callers.
	BizName string
	Codec   Codec
	// Decoders adds algorithms for reading frames written by other compressors.
	// The selected Codec is installed automatically. Duplicate types are rejected.
	Decoders []Decoder
	// Threshold is the minimum input size eligible for compression; zero disables it.
	Threshold int
	// MinReductionRatio is the minimum size reduction relative to the original data,
	// including the frame header in the compressed size. Valid values are [0, 1);
	// zero still requires the complete compressed frame to be smaller than the input.
	MinReductionRatio float64
	// MaxDecodedSize limits output, including uncompressed frames. Zero uses
	// DefaultMaxDecodedSize. A frame above the limit can be encoded but not decoded.
	MaxDecodedSize int
	// CompressCallback runs synchronously and may be called concurrently.
	CompressCallback func(context.Context, CompressionStats)
}

// DefaultConfig returns independent configuration with conservative policy
// defaults. Tune the thresholds against representative workloads.
func DefaultConfig() CompressorConfig {
	return CompressorConfig{
		Threshold:         DefaultThreshold,
		MinReductionRatio: DefaultMinReductionRatio,
		MaxDecodedSize:    DefaultMaxDecodedSize,
	}
}

type NoCompressReason string

const (
	NoCompressEmptyData             NoCompressReason = "emptyData"
	NoCompressBelowThreshold        NoCompressReason = "belowThreshold"
	NoCompressExpansion             NoCompressReason = "expansion"
	NoCompressInsufficientReduction NoCompressReason = "insufficientReduction"
	NoCompressFailed                NoCompressReason = "compressFailed"
)

// CompressionResult contains the result of a compression calling.
// Field Data always contains a complete frame, including on compression failure.
// Callers may store Data despite Err, or inspect Err to reject fallback.
type CompressionResult struct {
	Data   []byte
	Type   AlgType
	Reason NoCompressReason
	Err    error
}

type CompressionStats struct {
	BizName string
	// Algorithm is the configured codec, even when compression is skipped or fails.
	Algorithm   AlgType
	OriginalLen int
	ResultLen   int
	Reason      NoCompressReason
	Err         error
}

func NewCompressor(cfg CompressorConfig) (Compressor, error) {
	if cfg.Threshold < 0 {
		return nil, fmt.Errorf("%w: Threshold must be >= 0, got %d", ErrInvalidConfig, cfg.Threshold)
	}
	if cfg.MaxDecodedSize < 0 {
		return nil, fmt.Errorf("%w: MaxDecodedSize must be >= 0, got %d", ErrInvalidConfig, cfg.MaxDecodedSize)
	}
	if math.IsNaN(cfg.MinReductionRatio) || math.IsInf(cfg.MinReductionRatio, 0) || cfg.MinReductionRatio < 0 || cfg.MinReductionRatio >= 1 {
		return nil, fmt.Errorf("%w: MinReductionRatio must be finite and within [0, 1), got %g", ErrInvalidConfig, cfg.MinReductionRatio)
	}
	if cfg.Codec == nil {
		cfg.Codec = defaultGzipCodec
	}
	if nilInterface(cfg.Codec) {
		return nil, fmt.Errorf("%w: nil codec", ErrInvalidConfig)
	}
	if cfg.MaxDecodedSize == 0 {
		cfg.MaxDecodedSize = DefaultMaxDecodedSize
	}
	p := &compressorImpl{cfg: cfg, decoders: make(map[AlgType]Decoder)}
	for _, d := range append([]Decoder{cfg.Codec}, cfg.Decoders...) {
		if nilInterface(d) {
			return nil, fmt.Errorf("%w: nil decoder", ErrInvalidConfig)
		}
		typ := d.Type()
		if !isSupportedAlg(typ) {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedAlgorithm, typ)
		}
		if _, exists := p.decoders[typ]; exists {
			return nil, fmt.Errorf("%w: duplicate decoder %s", ErrInvalidConfig, typ)
		}
		p.decoders[typ] = d
	}
	p.cfg.Decoders = nil
	return p, nil
}

func nilInterface(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

type compressorImpl struct {
	cfg      CompressorConfig
	decoders map[AlgType]Decoder
}

func (p *compressorImpl) Compress(ctx context.Context, data []byte) (result CompressionResult) {
	alg := p.cfg.Codec.Type()
	if p.cfg.CompressCallback != nil {
		defer func() {
			stats := CompressionStats{
				BizName:     p.cfg.BizName,
				Algorithm:   alg,
				OriginalLen: len(data),
				ResultLen:   len(result.Data),
				Reason:      result.Reason,
				Err:         result.Err,
			}
			p.cfg.CompressCallback(ctx, stats)
		}()
	}
	if len(data) == 0 {
		return uncompressedResult(data, NoCompressEmptyData, nil)
	}
	if len(data) < p.cfg.Threshold {
		return uncompressedResult(data, NoCompressBelowThreshold, nil)
	}
	header := makeHeader(alg, len(data))
	out, err := p.cfg.Codec.Compress(header, data)
	if err == nil {
		_, resultAlg, decodedSize, frameErr := unpackFrame(out)
		if frameErr != nil || resultAlg != alg || decodedSize != uint64(len(data)) {
			err = fmt.Errorf("codec %s did not preserve frame header", alg)
		}
	}
	if err != nil {
		return uncompressedResult(data, NoCompressFailed, fmt.Errorf("compress %s: %w", alg, err))
	}
	reducedBytes := len(data) - len(out)
	if reducedBytes < 0 {
		return uncompressedResult(data, NoCompressExpansion, nil)
	}
	reductionRatio := float64(reducedBytes) / float64(len(data))
	if reducedBytes == 0 || reductionRatio < p.cfg.MinReductionRatio {
		return uncompressedResult(data, NoCompressInsufficientReduction, nil)
	}
	return CompressionResult{Data: out, Type: alg}
}

func makeHeader(alg AlgType, size int) []byte {
	h := make([]byte, frameHeaderSize)
	copy(h, frameMagic)
	h[frameVersionOffset], h[frameAlgorithmOffset] = frameVersion, byte(alg)
	binary.BigEndian.PutUint64(h[frameSizeOffset:], uint64(size))
	return h
}

func uncompressedResult(data []byte, reason NoCompressReason, err error) CompressionResult {
	return CompressionResult{Data: append(makeHeader(TypeNone, len(data)), data...), Type: TypeNone, Reason: reason, Err: err}
}

func (p *compressorImpl) Decompress(_ context.Context, data []byte) ([]byte, AlgType, error) {
	payload, alg, size, err := unpackFrame(data)
	if err != nil {
		return nil, alg, err
	}
	if size > uint64(p.cfg.MaxDecodedSize) {
		return nil, alg, ErrDecodedTooLarge
	}
	if alg == TypeNone {
		if uint64(len(payload)) != size {
			return nil, alg, ErrInvalidFrame
		}
		return bytes.Clone(payload), alg, nil
	}
	d, ok := p.decoders[alg]
	if !ok {
		return nil, alg, fmt.Errorf("%w: %s", ErrDecoderUnavailable, alg)
	}
	out, err := d.Decompress(payload, p.cfg.MaxDecodedSize)
	if err != nil {
		return nil, alg, fmt.Errorf("decompress %s: %w", alg, err)
	}
	if len(out) > p.cfg.MaxDecodedSize {
		return nil, alg, ErrDecodedTooLarge
	}
	if uint64(len(out)) != size {
		return nil, alg, ErrInvalidFrame
	}
	return out, alg, nil
}

func unpackFrame(data []byte) (payload []byte, alg AlgType, decodedSize uint64, err error) {
	if len(data) < frameHeaderSize || string(data[:frameVersionOffset]) != frameMagic || data[frameVersionOffset] != frameVersion {
		return nil, TypeUnknown, 0, ErrInvalidFrame
	}
	alg = AlgType(data[frameAlgorithmOffset])
	if alg != TypeNone && !isSupportedAlg(alg) {
		return nil, alg, 0, ErrUnsupportedAlgorithm
	}
	decodedSize = binary.BigEndian.Uint64(data[frameSizeOffset:frameHeaderSize])
	return data[frameHeaderSize:], alg, decodedSize, nil
}
