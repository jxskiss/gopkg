package compress

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jxskiss/gopkg/v2/internal"
)

const (
	DefaultMinSaving = 0.05
	DefaultThreshold = 5 * 1024 // 5KB
)

// DefaultCompressor is the default Compressor, using gzip algorithm,
// compress level is gzip.DefaultCompression.
//
// It's recommended to use zstd for better performance.
var DefaultCompressor Compressor

func init() {
	DefaultCompressor = NewCompressor(CompressorConfig{
		BizName: "default",
		Alg:     defaultGzipAlg,
	})
}

// Compressor 提供统一的压缩/解压缩接口，压缩/解压缩算法的实现必须是并发安全的。
type Compressor interface {
	// Compress 压缩数据，返回压缩后的数据和压缩类型。
	// 注意 Compress 方法返回的 compressType 为 TypeNoCompress 时，返回的 result
	// 也会带有 header, 跟原始 data 不一致，使用者需要保存返回的 result 数据。
	// 如果返回的 compressType 是 TypeNoCompress, noCompressReason 是压缩失败的原因。
	// 如果在压缩失败情况下要中断业务流程，可以检查 noCompressReason == ReasonCompressFailed.
	Compress(ctx context.Context, data []byte) (result []byte, compressType AlgType, noCompressReason string)

	// Decompress 解压缩数据，返回解压缩后的数据和压缩类型。
	// 解压缩不依赖 Compress 方法返回的 compressType，但传入的 data 必须是
	// Compress 方法返回的 result 数据。
	Decompress(ctx context.Context, data []byte) (result []byte, compressType AlgType, err error)
}

type CompressorConfig struct {
	Alg       CompressionAlg // 压缩算法实现
	BizName   string         // 业务名称，用于在打点和错误日志中标识不同的业务场景
	Threshold int            // 压缩阈值，当数据长度小于 threshold 时不压缩
	MinSaving float64        // 压缩收益阈值，当压缩收益小于 minSaving 时不压缩

	// CompressCallback 用于 Compress 方法结束时回调，可用于观测压缩指标, optional
	CompressCallback func(ctx context.Context, info CompressionInfo)

	// ErrorLogger 用于记录压缩错误日志, optional
	ErrorLogger func(ctx context.Context, err error, msg string)

	TreatNoHeaderAsNoCompress bool
}

type CompressionInfo struct {
	Config           *CompressorConfig
	OriginalLen      int
	ResultLen        int
	CompressType     AlgType
	NoCompressReason string
}

func NewCompressor(cfg CompressorConfig) Compressor {
	compressor := &compressorImpl{
		CompressorConfig: cfg,
	}
	compressor.setup()
	return compressor
}

type compressorImpl struct {
	CompressorConfig
	header []byte
}

func (p *compressorImpl) setup() {
	if p.Alg == nil {
		p.Alg = DefaultCompressor.(*compressorImpl).Alg
	}
	if p.Threshold <= 0 {
		p.Threshold = DefaultThreshold
	}
	if p.MinSaving <= 0 {
		p.MinSaving = DefaultMinSaving
	}
	if p.ErrorLogger == nil {
		p.ErrorLogger = internal.DefaultLoggerError
	}
	p.header = []byte{byte(p.Alg.Type()), headerByte2}
}

func (p *compressorImpl) Compress(ctx context.Context, data []byte) (result []byte, compressType AlgType, noCompressReason string) {
	if p.CompressCallback != nil {
		defer func() {
			p.CompressCallback(ctx, CompressionInfo{
				Config:           &p.CompressorConfig,
				OriginalLen:      len(data),
				ResultLen:        len(result),
				CompressType:     compressType,
				NoCompressReason: noCompressReason,
			})
		}()
	}

	if len(data) == 0 {
		return data, TypeNoCompress, ReasonEmptyData
	}
	if len(data) < p.Threshold {
		return p.noCompress(data, ReasonBelowThreshold)
	}

	result, err := p.Alg.Compress(p.header, data)
	if err != nil {
		// Compressing failed is a very rare case,
		// write an error log to avoid silent failure.
		p.ErrorLogger(ctx, err, fmt.Sprintf("compressing failed: bizName= %s, algType= %s", p.BizName, p.Alg.Type().String()))
		return p.noCompress(data, ReasonCompressFailed)
	}

	saving := 1 - float64(len(result))/float64(len(data))
	if saving >= p.MinSaving {
		return result, p.Alg.Type(), ""
	}

	noCompressReason = ReasonSavingTooSmall
	if saving < 0 {
		noCompressReason = ReasonSavingNegative
	}
	return p.noCompress(data, noCompressReason)
}

func (p *compressorImpl) noCompress(data []byte, reason string) (result []byte, compressType AlgType, noCompressReason string) {
	result = append(headerNoCompress, data...)
	return result, TypeNoCompress, reason
}

func (p *compressorImpl) Decompress(_ context.Context, data []byte) (result []byte, compressType AlgType, err error) {
	if len(data) < 2 {
		return data, TypeNoCompress, nil
	}
	if bytes.HasPrefix(data, headerNoCompress) {
		return data[2:], TypeNoCompress, nil
	}
	if data[0] == byte(TypeGzip) && data[1] == headerByte2 {
		result, err = gzipDecompressFunc(data[2:])
		return result, TypeGzip, err
	}
	if data[0] == byte(TypeZstd) && data[1] == headerByte2 {
		return decompressZstd(data[2:])
	}
	if bytes.HasPrefix(data, gzipMagicNumber) {
		result, err = gzipDecompressFunc(data)
		return result, TypeGzip, err
	}
	if bytes.HasPrefix(data, zstdMagicNumber) {
		return decompressZstd(data)
	}
	if p.TreatNoHeaderAsNoCompress {
		return data, TypeNoCompress, nil
	}
	// 为防止误用，不带 header 的数据默认返回错误以提醒使用者保存带 header 的数据
	return data, TypeUnknown, fmt.Errorf("unknown compression type %s", AlgType(data[0]))
}
