package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"sync"
)

var defaultGzipAlg = NewGzipCompressor(gzip.DefaultCompression)

func init() {
	ProvideDecompressor(defaultGzipAlg)
}

// NewGzipCompressor creates a new gzip CompressionAlg instance.
// level specifies the compression level, see package gzip for valid values,
// the default value is gzip.DefaultCompression.
//
// It's recommended to use zstd for better performance.
func NewGzipCompressor(level int) CompressionAlg {
	if level == 0 {
		level = gzip.DefaultCompression
	}
	return &GzipCompressor{
		level: level,
	}
}

var gzipReaderPool = sync.Pool{
	New: func() interface{} { return new(gzip.Reader) },
}

type GzipCompressor struct {
	level      int
	writerPool sync.Pool
}

func (p *GzipCompressor) Type() AlgType {
	return TypeGzip
}

func (p *GzipCompressor) CompressionLevel() int {
	return p.level
}

func (p *GzipCompressor) Compress(dst []byte, data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(dst)
	gw, err := p.getWriter(buf)
	if err != nil {
		return nil, err
	}
	defer p.writerPool.Put(gw)

	_, err = gw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("gzip compress failed: %w", err)
	}
	if err = gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip compress failed: %w", err)
	}
	return buf.Bytes(), nil
}

func (p *GzipCompressor) getWriter(buf *bytes.Buffer) (*gzip.Writer, error) {
	var err error
	var gw *gzip.Writer
	if v := p.writerPool.Get(); v != nil {
		gw = v.(*gzip.Writer)
		gw.Reset(buf)
	} else {
		gw, err = gzip.NewWriterLevel(buf, p.level)
		if err != nil {
			return nil, fmt.Errorf("gzip.NewWriterLevel failed: %w", err)
		}
	}
	return gw, nil
}

func (p *GzipCompressor) Decompress(data []byte) ([]byte, error) {
	gr := gzipReaderPool.Get().(*gzip.Reader)
	err := gr.Reset(bytes.NewReader(data))
	if err != nil {
		gr, err = gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gzip.NewReader failed: %w", err)
		}
	}
	defer gzipReaderPool.Put(gr)

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(gr); err != nil {
		return nil, fmt.Errorf("gzip decompress failed: %w", err)
	}
	return buf.Bytes(), nil
}
