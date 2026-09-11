package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

var defaultGzipCodec = &GzipCodec{level: gzip.DefaultCompression}

// NewGzipCodec accepts the standard library's gzip levels, including
// NoCompression (0), DefaultCompression (-1), and HuffmanOnly (-2).
func NewGzipCodec(level int) (*GzipCodec, error) {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return nil, fmt.Errorf("%w: gzip level %d", ErrInvalidConfig, level)
	}
	return &GzipCodec{level: level}, nil
}

// GzipCodec must not be copied after first use. Its zero value uses NoCompression.
type GzipCodec struct {
	level   int
	writers sync.Pool
}

func (p *GzipCodec) Type() AlgType         { return TypeGzip }
func (p *GzipCodec) CompressionLevel() int { return p.level }

type gzipWriter struct {
	buf    bytes.Buffer
	writer *gzip.Writer
}

func (p *GzipCodec) Compress(dst, data []byte) ([]byte, error) {
	var w *gzipWriter
	if v := p.writers.Get(); v != nil {
		w = v.(*gzipWriter)
	} else {
		w = new(gzipWriter)
	}
	w.buf = *bytes.NewBuffer(dst)
	if w.writer == nil {
		var err error
		w.writer, err = gzip.NewWriterLevel(&w.buf, p.level)
		if err != nil {
			return nil, err
		}
	} else {
		w.writer.Reset(&w.buf)
	}
	defer func() {
		// The writer retains only this wrapper, never the caller's output buffer.
		w.buf = bytes.Buffer{}
		p.writers.Put(w)
	}()
	if _, err := w.writer.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := w.writer.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}
	return w.buf.Bytes(), nil
}

type gzipReader struct {
	src    bytes.Reader
	reader gzip.Reader
}

var gzipReaders = sync.Pool{New: func() interface{} { return new(gzipReader) }}

func (p *GzipCodec) Decompress(data []byte, maxDecodedSize int) ([]byte, error) {
	if maxDecodedSize <= 0 {
		return nil, fmt.Errorf("%w: max decoded size must be positive", ErrInvalidConfig)
	}
	r := gzipReaders.Get().(*gzipReader)
	defer func() {
		r.src.Reset(nil)
		r.reader.Header = gzip.Header{}
		gzipReaders.Put(r)
	}()
	r.src.Reset(data)
	if err := r.reader.Reset(&r.src); err != nil {
		return nil, fmt.Errorf("gzip header: %w", err)
	}
	defer r.reader.Close()
	out, err := io.ReadAll(io.LimitReader(&r.reader, int64(maxDecodedSize)))
	if err != nil {
		return nil, fmt.Errorf("gzip read: %w", err)
	}
	// Probe EOF even at the exact limit to validate the trailer and detect excess
	// output without computing maxDecodedSize+1, which could overflow.
	var extra [1]byte
	n, err := io.ReadFull(&r.reader, extra[:])
	if n != 0 {
		return nil, ErrDecodedTooLarge
	}
	if err != io.EOF {
		return nil, fmt.Errorf("gzip trailer: %w", err)
	}
	return out, nil
}
