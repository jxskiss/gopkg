package compress

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"testing"
)

func newTestCompressor(t testing.TB, cfg CompressorConfig) Compressor {
	t.Helper()
	c, err := NewCompressor(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func makeUncompressedFrame(data []byte) []byte {
	return append(makeHeader(TypeNone, len(data)), data...)
}

type stubCodec struct {
	kind   AlgType
	encode func([]byte, []byte) ([]byte, error)
	decode func([]byte, int) ([]byte, error)
}

func (c *stubCodec) Type() AlgType                                    { return c.kind }
func (c *stubCodec) Compress(dst, src []byte) ([]byte, error)         { return c.encode(dst, src) }
func (c *stubCodec) Decompress(src []byte, limit int) ([]byte, error) { return c.decode(src, limit) }

func TestCompressorRoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		data   []byte
		cfg    CompressorConfig
		typ    AlgType
		reason NoCompressReason
	}{
		{"nil", nil, DefaultConfig(), TypeNone, NoCompressEmptyData},
		{"empty", []byte{}, DefaultConfig(), TypeNone, NoCompressEmptyData},
		{"short", []byte("hello"), DefaultConfig(), TypeNone, NoCompressBelowThreshold},
		{"chinese", []byte("你好，世界"), DefaultConfig(), TypeNone, NoCompressBelowThreshold},
		{"magic plaintext", []byte{'0', 1, 0x1f, 0x8b, 0x08}, DefaultConfig(), TypeNone, NoCompressBelowThreshold},
		{"gzip", bytes.Repeat([]byte("hello world"), 1024), DefaultConfig(), TypeGzip, ""},
		{"zero thresholds", bytes.Repeat([]byte("a"), 100), CompressorConfig{}, TypeGzip, ""},
		{"expansion", []byte("hello"), CompressorConfig{}, TypeNone, NoCompressExpansion},
		{"insufficient", bytes.Repeat([]byte("a"), 100), CompressorConfig{MinReductionRatio: math.Nextafter(1, 0)}, TypeNone, NoCompressInsufficientReduction},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCompressor(t, tc.cfg)
			result := c.Compress(context.Background(), tc.data)
			if result.Err != nil || result.Type != tc.typ || result.Reason != tc.reason {
				t.Fatalf("unexpected result: type=%s reason=%s err=%v", result.Type, result.Reason, result.Err)
			}
			out, typ, err := c.Decompress(context.Background(), result.Data)
			if err != nil || typ != tc.typ || !bytes.Equal(out, tc.data) {
				t.Fatalf("round trip: type=%s err=%v", typ, err)
			}
		})
	}
}

func TestCompressorConfig(t *testing.T) {
	var nilCodec *GzipCodec
	cases := []struct {
		name string
		cfg  CompressorConfig
		want error
	}{
		{"negative threshold", CompressorConfig{Threshold: -1}, ErrInvalidConfig},
		{"negative limit", CompressorConfig{MaxDecodedSize: -1}, ErrInvalidConfig},
		{"negative reduction", CompressorConfig{MinReductionRatio: -0.1}, ErrInvalidConfig},
		{"full reduction", CompressorConfig{MinReductionRatio: 1}, ErrInvalidConfig},
		{"excess reduction", CompressorConfig{MinReductionRatio: 1.1}, ErrInvalidConfig},
		{"nan", CompressorConfig{MinReductionRatio: math.NaN()}, ErrInvalidConfig},
		{"inf", CompressorConfig{MinReductionRatio: math.Inf(1)}, ErrInvalidConfig},
		{"negative inf", CompressorConfig{MinReductionRatio: math.Inf(-1)}, ErrInvalidConfig},
		{"typed nil codec", CompressorConfig{Codec: nilCodec}, ErrInvalidConfig},
		{"nil decoder", CompressorConfig{Decoders: []Decoder{nil}}, ErrInvalidConfig},
		{"typed nil decoder", CompressorConfig{Decoders: []Decoder{nilCodec}}, ErrInvalidConfig},
		{"duplicate", CompressorConfig{Decoders: []Decoder{defaultGzipCodec}}, ErrInvalidConfig},
		{"none encoder", CompressorConfig{Codec: &stubCodec{kind: TypeNone}}, ErrUnsupportedAlgorithm},
		{"unknown encoder", CompressorConfig{Codec: &stubCodec{kind: 42}}, ErrUnsupportedAlgorithm},
		{"unknown decoder", CompressorConfig{Decoders: []Decoder{&stubCodec{kind: 42}}}, ErrUnsupportedAlgorithm},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewCompressor(tc.cfg)
			if c != nil || !errors.Is(err, tc.want) {
				t.Fatalf("got %v, %v", c, err)
			}
		})
	}
}

func TestCompressorPolicyBoundaries(t *testing.T) {
	calls := 0
	codec := &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) {
		calls++
		return append(dst, src[:50]...), nil
	}}
	input := bytes.Repeat([]byte("x"), 100)
	ratio := float64(100-(50+frameHeaderSize)) / 100
	cases := []struct {
		name           string
		threshold      int
		reductionRatio float64
		want           AlgType
	}{
		{"threshold below", 101, 0, TypeNone},
		{"threshold equal", 100, 0, TypeGzip},
		{"reduction equal", 0, ratio, TypeGzip},
		{"reduction below", 0, math.Nextafter(ratio, 1), TypeNone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCompressor(t, CompressorConfig{Codec: codec, Threshold: tc.threshold, MinReductionRatio: tc.reductionRatio})
			result := c.Compress(context.Background(), input)
			if result.Type != tc.want || result.Err != nil {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
	if calls != len(cases)-1 {
		t.Fatalf("encode calls = %d", calls)
	}
	c := newTestCompressor(t, CompressorConfig{Codec: codec})
	for _, tc := range []struct {
		payloadSize int
		reason      NoCompressReason
	}{
		{100 - frameHeaderSize, NoCompressInsufficientReduction},
		{100 - frameHeaderSize + 1, NoCompressExpansion},
	} {
		codec.encode = func(dst, src []byte) ([]byte, error) { return append(dst, src[:tc.payloadSize]...), nil }
		result := c.Compress(context.Background(), input)
		if result.Type != TypeNone || result.Reason != tc.reason || result.Err != nil {
			t.Fatalf("frame size %d: %+v", tc.payloadSize+frameHeaderSize, result)
		}
	}
}

func TestCompressorFallbackAndObserver(t *testing.T) {
	failure := errors.New("encoder failed")
	codec := &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) {
		dst[0] = 0
		return nil, failure
	}}
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "trace")
	var observed CompressionStats
	var calls int
	cfg := CompressorConfig{Codec: codec, BizName: "test", CompressCallback: func(got context.Context, info CompressionStats) {
		if got != ctx {
			t.Error("context lost")
		}
		calls++
		observed = info
		info.BizName = "changed"
	}}
	c := newTestCompressor(t, cfg)
	cfg.Codec = nil
	cfg.BizName = "outside mutation"
	input := []byte("hello")
	result := c.Compress(ctx, input)
	if result.Type != TypeNone || result.Reason != NoCompressFailed || !errors.Is(result.Err, failure) {
		t.Fatalf("%+v", result)
	}
	out, _, err := c.Decompress(ctx, result.Data)
	if err != nil || !bytes.Equal(out, input) {
		t.Fatalf("fallback: %v", err)
	}
	if calls != 1 || observed.BizName != "test" || observed.Algorithm != TypeGzip ||
		observed.OriginalLen != len(input) || observed.ResultLen != len(result.Data) ||
		observed.Reason != result.Reason || !errors.Is(observed.Err, failure) {
		t.Fatalf("%+v", observed)
	}
	codec.encode = func(dst, src []byte) ([]byte, error) { return []byte("broken prefix"), nil }
	if got := c.Compress(ctx, input); got.Reason != NoCompressFailed || got.Err == nil {
		t.Fatal("invalid codec output accepted")
	}
}

func TestCompressorFrameFormat(t *testing.T) {
	c := newTestCompressor(t, CompressorConfig{DisableUnframed: true})
	input := []byte("abc")
	want := []byte{
		0x13, 0x39, 0xce, 0x30, 0xc0, 0x5b, 0x98, 0xc4,
		0x4e, 0x13, 0x21, 0x15, 0x35, 0x55, 0xc4, 0xb8,
		1, '0', 0, 0, 0, 0, 0, 0, 0, 3, 'a', 'b', 'c',
	}
	result := c.Compress(context.Background(), input)
	if result.Err != nil || !bytes.Equal(result.Data, want) {
		t.Fatalf("frame = %x, err = %v", result.Data, result.Err)
	}
	out, typ, err := c.Decompress(context.Background(), want)
	if err != nil || typ != TypeNone || !bytes.Equal(out, input) {
		t.Fatalf("decode fixed frame: type=%s err=%v", typ, err)
	}
}

func TestCompressorDefaultUnframed(t *testing.T) {
	c := newTestCompressor(t, CompressorConfig{MaxDecodedSize: 3})
	input := []byte("abc")
	out, typ, err := c.Decompress(context.Background(), input)
	if err != nil || typ != TypeNone || !bytes.Equal(out, input) {
		t.Fatalf("raw: %q %v %v", out, typ, err)
	}
	if &out[0] != &input[0] {
		t.Fatal("unframed decode unexpectedly copied")
	}
	out[0] = 'x'
	if string(input) != "xbc" {
		t.Fatal("decoded output does not share input")
	}
	for _, input := range [][]byte{[]byte("ab"), []byte("abc"), []byte("abcd")} {
		if out, typ, err := c.Decompress(context.Background(), input); err != nil || typ != TypeNone || !bytes.Equal(out, input) {
			t.Fatalf("raw: %q %v %v", out, typ, err)
		}
	}
}

func TestCompressorUnframedBorrowsInput(t *testing.T) {
	c := newTestCompressor(t, DefaultConfig())
	input := []byte("abc")
	result := c.Compress(context.Background(), input)
	if result.Err != nil || result.Type != TypeNone || result.Reason != NoCompressBelowThreshold ||
		!bytes.Equal(result.Data, input) || &result.Data[0] != &input[0] {
		t.Fatalf("expected original input: %+v", result)
	}
	input[0] = 'x'
	if result.Data[0] != 'x' {
		t.Fatal("unframed output unexpectedly copied")
	}
}

func TestCompressorTypeNoneBorrowsPayload(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		c := newTestCompressor(t, CompressorConfig{DisableUnframed: disabled})
		frame := makeUncompressedFrame([]byte("abc"))
		saved := bytes.Clone(frame)
		out, typ, err := c.Decompress(context.Background(), frame)
		if err != nil || typ != TypeNone || string(out) != "abc" {
			t.Fatalf("decode: %q %v %v", out, typ, err)
		}
		if !bytes.Equal(frame, saved) {
			t.Fatal("decoder modified frame")
		}
		if &out[0] != &frame[frameHeaderSize] {
			t.Fatal("TypeNone payload unexpectedly copied")
		}
		out[0] = 'x'
		if frame[frameHeaderSize] != 'x' {
			t.Fatal("decoded payload does not share frame")
		}
	}
}

func TestCompressorUnframedMatrix(t *testing.T) {
	ctx := context.Background()
	magic := []byte(frameMagic)
	rawGzip, err := defaultGzipCodec.Compress(nil, []byte("plain gzip stream"))
	if err != nil {
		t.Fatal(err)
	}
	cases := [][]byte{nil, {}, {0}, {0xff, 0, 1}, rawGzip, {0x28, 0xb5, 0x2f, 0xfd}}
	for n := 1; n < len(magic); n++ {
		cases = append(cases, bytes.Clone(magic[:n]))
	}
	compatible := newTestCompressor(t, CompressorConfig{})
	strict := newTestCompressor(t, CompressorConfig{DisableUnframed: true})
	for i, input := range cases {
		out, typ, err := compatible.Decompress(ctx, input)
		if err != nil || typ != TypeNone || !bytes.Equal(out, input) {
			t.Fatalf("compatible case %d: %x %v %v", i, out, typ, err)
		}
		if input == nil && out != nil {
			t.Fatal("nil input became non-nil")
		}
		if input != nil && out == nil {
			t.Fatal("non-nil empty input became nil")
		}
		strictOut, strictTyp, strictErr := strict.Decompress(ctx, input)
		if strictOut != nil || strictTyp != TypeUnknown || !errors.Is(strictErr, ErrInvalidFrame) {
			t.Fatalf("strict case %d: %x %v %v", i, strictOut, strictTyp, strictErr)
		}
	}
}

func TestCompressorDefaultUnframedRejectsDamagedFrame(t *testing.T) {
	c := newTestCompressor(t, CompressorConfig{})
	frame := makeUncompressedFrame([]byte("abc"))
	badVersion := bytes.Clone(frame)
	badVersion[16]++
	unknown := bytes.Clone(frame)
	unknown[frameAlgorithmOffset] = 42
	missing := bytes.Clone(frame)
	missing[frameAlgorithmOffset] = byte(TypeZstd)
	gzipFrame := c.Compress(context.Background(), bytes.Repeat([]byte("a"), 1024)).Data
	gzipFrame[len(gzipFrame)-1] ^= 1
	cases := []struct {
		data []byte
		want error
	}{
		{frame[:16], ErrInvalidFrame},
		{frame[:25], ErrInvalidFrame},
		{frame[:len(frame)-1], ErrInvalidFrame},
		{badVersion, ErrInvalidFrame},
		{unknown, ErrUnsupportedAlgorithm},
		{missing, ErrDecoderUnavailable},
		{gzipFrame, nil},
	}
	for _, tc := range cases {
		out, _, err := c.Decompress(context.Background(), tc.data)
		if out != nil || err == nil || tc.want != nil && !errors.Is(err, tc.want) {
			t.Fatalf("accepted damaged frame: %x %v", tc.data, err)
		}
	}
}

func TestCompressorFallbackCallbacks(t *testing.T) {
	failure := errors.New("failed")
	magicInput := append([]byte(frameMagic), 'x')
	cases := []struct {
		name  string
		input []byte
		cfg   CompressorConfig
	}{
		{"empty", nil, CompressorConfig{}},
		{"below threshold", []byte("abc"), CompressorConfig{Threshold: 4}},
		{"magic collision", magicInput, CompressorConfig{Threshold: len(magicInput) + 1}},
		{"expansion", bytes.Repeat([]byte("x"), 100), CompressorConfig{Codec: &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) { return append(dst, src...), nil }}}},
		{"insufficient", bytes.Repeat([]byte("x"), 100), CompressorConfig{Codec: &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) { return append(dst, src[:1]...), nil }}, MinReductionRatio: math.Nextafter(1, 0)}},
		{"failure", bytes.Repeat([]byte("x"), 100), CompressorConfig{Codec: &stubCodec{kind: TypeGzip, encode: func([]byte, []byte) ([]byte, error) { return nil, failure }}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stats CompressionStats
			tc.cfg.CompressCallback = func(_ context.Context, got CompressionStats) { stats = got }
			result := newTestCompressor(t, tc.cfg).Compress(context.Background(), tc.input)
			if stats.OriginalLen != len(tc.input) || stats.ResultLen != len(result.Data) || stats.Reason != result.Reason || !errors.Is(stats.Err, result.Err) {
				t.Fatalf("result=%+v stats=%+v", result, stats)
			}
		})
	}
}

func TestCompressorUnframedFallbacksBorrowInput(t *testing.T) {
	failure := errors.New("failed")
	input := bytes.Repeat([]byte("x"), 100)
	cases := []struct {
		name    string
		codec   *stubCodec
		ratio   float64
		reason  NoCompressReason
		wantErr error
	}{
		{"expansion", &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) { return append(dst, src...), nil }}, 0, NoCompressExpansion, nil},
		{"insufficient", &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) { return append(dst, src[:1]...), nil }}, math.Nextafter(1, 0), NoCompressInsufficientReduction, nil},
		{"error", &stubCodec{kind: TypeGzip, encode: func([]byte, []byte) ([]byte, error) { return nil, failure }}, 0, NoCompressFailed, failure},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCompressor(t, CompressorConfig{Codec: tc.codec, MinReductionRatio: tc.ratio})
			result := c.Compress(context.Background(), input)
			if result.Type != TypeNone || result.Reason != tc.reason || !errors.Is(result.Err, tc.wantErr) || &result.Data[0] != &input[0] {
				t.Fatalf("fallback: %+v", result)
			}
		})
	}
}

func TestCompressorMagicCollisionEscaped(t *testing.T) {
	magicInput := append([]byte(frameMagic), bytes.Repeat([]byte("x"), 84)...)
	failure := errors.New("failed")
	cases := []CompressorConfig{
		{Threshold: len(magicInput) + 1},
		{Codec: &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) { return append(dst, src...), nil }}},
		{Codec: &stubCodec{kind: TypeGzip, encode: func(dst, src []byte) ([]byte, error) { return append(dst, src[:1]...), nil }}, MinReductionRatio: math.Nextafter(1, 0)},
		{Codec: &stubCodec{kind: TypeGzip, encode: func([]byte, []byte) ([]byte, error) { return nil, failure }}},
	}
	for i, cfg := range cases {
		c := newTestCompressor(t, cfg)
		result := c.Compress(context.Background(), magicInput)
		if result.Type != TypeNone || len(result.Data) != frameHeaderSize+len(magicInput) || &result.Data[0] == &magicInput[0] {
			t.Fatalf("case %d not escaped: %+v", i, result)
		}
		out, typ, err := c.Decompress(context.Background(), result.Data)
		if err != nil || typ != TypeNone || !bytes.Equal(out, magicInput) {
			t.Fatalf("case %d round trip: %v %v", i, typ, err)
		}
	}
}

func TestCompressorEmptyAlwaysFramed(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		c := newTestCompressor(t, CompressorConfig{DisableUnframed: disabled})
		for _, input := range [][]byte{nil, {}} {
			result := c.Compress(context.Background(), input)
			if result.Type != TypeNone || result.Reason != NoCompressEmptyData || len(result.Data) != frameHeaderSize {
				t.Fatalf("empty result: %+v", result)
			}
		}
	}
}

func TestCompressorStrictFrames(t *testing.T) {
	c := newTestCompressor(t, CompressorConfig{DisableUnframed: true})
	ctx := context.Background()
	input := bytes.Repeat([]byte("a"), 1024)
	valid := c.Compress(ctx, input).Data
	raw, _ := defaultGzipCodec.Compress(nil, input)
	invalid := [][]byte{nil, {}, []byte("a"), []byte("plaintext"), raw, {'0', 1, 'x'}, {0x28, 0xb5, 0x2f, 0xfd}}
	for i := 0; i < len(valid); i++ {
		invalid = append(invalid, bytes.Clone(valid[:i]))
	}
	version := bytes.Clone(valid)
	version[frameVersionOffset]++
	invalid = append(invalid, version)
	length := bytes.Clone(valid)
	binary.BigEndian.PutUint64(length[frameSizeOffset:], 1023)
	invalid = append(invalid, length)
	trailer := bytes.Clone(valid)
	trailer[len(trailer)-8] ^= 1
	invalid = append(invalid, trailer)
	invalid = append(invalid, append(bytes.Clone(valid), 1))
	plain := makeUncompressedFrame([]byte("abc"))
	invalid = append(invalid, plain[:len(plain)-1], append(bytes.Clone(plain), 1))
	for i, data := range invalid {
		out, _, err := c.Decompress(ctx, data)
		if err == nil || out != nil {
			t.Fatalf("case %d accepted invalid input", i)
		}
	}
	unknown := bytes.Clone(valid)
	unknown[frameAlgorithmOffset] = 42
	if _, typ, err := c.Decompress(ctx, unknown); typ != 42 || !errors.Is(err, ErrUnsupportedAlgorithm) {
		t.Fatalf("unknown: %s %v", typ, err)
	}
	missing := bytes.Clone(valid)
	missing[frameAlgorithmOffset] = byte(TypeZstd)
	if _, typ, err := c.Decompress(ctx, missing); typ != TypeZstd || !errors.Is(err, ErrDecoderUnavailable) {
		t.Fatalf("missing: %s %v", typ, err)
	}
}

func TestCompressorDecodeLimit(t *testing.T) {
	ctx := context.Background()
	input := bytes.Repeat([]byte("a"), 1024)
	for _, threshold := range []int{0, 2048} {
		encoder := newTestCompressor(t, CompressorConfig{Threshold: threshold, DisableUnframed: true})
		frame := encoder.Compress(ctx, input).Data
		for _, limit := range []int{1023, 1024, 1025} {
			c := newTestCompressor(t, CompressorConfig{MaxDecodedSize: limit})
			out, _, err := c.Decompress(ctx, frame)
			if threshold == 0 && limit < len(input) {
				if out != nil || !errors.Is(err, ErrDecodedTooLarge) {
					t.Fatalf("limit: %v", err)
				}
			} else if err != nil || !bytes.Equal(out, input) {
				t.Fatalf("within limit: %v", err)
			}
		}
	}
	frame := newTestCompressor(t, CompressorConfig{DisableUnframed: true}).Compress(ctx, input).Data
	binary.BigEndian.PutUint64(frame[frameSizeOffset:], 1)
	c := newTestCompressor(t, CompressorConfig{MaxDecodedSize: 32})
	if out, _, err := c.Decompress(ctx, frame); out != nil || !errors.Is(err, ErrDecodedTooLarge) {
		t.Fatalf("forged size: %v", err)
	}
	binary.BigEndian.PutUint64(frame[frameSizeOffset:], math.MaxUint64)
	if _, _, err := c.Decompress(ctx, frame); !errors.Is(err, ErrDecodedTooLarge) {
		t.Fatalf("overflow: %v", err)
	}
}

func TestCompressorUncompressedIgnoresDecodeLimit(t *testing.T) {
	ctx := context.Background()
	input := bytes.Repeat([]byte("x"), 1024)
	for _, disabled := range []bool{false, true} {
		c := newTestCompressor(t, CompressorConfig{MaxDecodedSize: 1, DisableUnframed: disabled})
		frame := makeUncompressedFrame(input)
		out, typ, err := c.Decompress(ctx, frame)
		if err != nil || typ != TypeNone || !bytes.Equal(out, input) || &out[0] != &frame[frameHeaderSize] {
			t.Fatalf("uncompressed frame: %v %v", typ, err)
		}
		if !disabled {
			out, typ, err = c.Decompress(ctx, input)
			if err != nil || typ != TypeNone || !bytes.Equal(out, input) || &out[0] != &input[0] {
				t.Fatalf("unframed: %v %v", typ, err)
			}
		}
		for _, size := range []uint64{1023, 1025, math.MaxUint64} {
			binary.BigEndian.PutUint64(frame[frameSizeOffset:], size)
			out, typ, err = c.Decompress(ctx, frame)
			if out != nil || typ != TypeNone || !errors.Is(err, ErrInvalidFrame) {
				t.Fatalf("invalid uncompressed length %d: %v %v", size, typ, err)
			}
		}
	}
}

func TestCompressorDecoderIsolation(t *testing.T) {
	input := bytes.Repeat([]byte("a"), 100)
	codec := &stubCodec{kind: TypeZstd, encode: func(dst, src []byte) ([]byte, error) { return append(dst, 1), nil },
		decode: func(src []byte, limit int) ([]byte, error) { return bytes.Clone(input), nil }}
	decoders := []Decoder{codec}
	reader := newTestCompressor(t, CompressorConfig{Decoders: decoders})
	writer := newTestCompressor(t, CompressorConfig{Codec: codec})
	decoders[0] = nil
	frame := writer.Compress(context.Background(), input).Data
	for _, c := range []Compressor{reader, writer} {
		out, typ, err := c.Decompress(context.Background(), frame)
		if err != nil || typ != TypeZstd || !bytes.Equal(out, input) {
			t.Fatalf("%s %v", typ, err)
		}
	}
	other := newTestCompressor(t, CompressorConfig{})
	if _, _, err := other.Decompress(context.Background(), frame); !errors.Is(err, ErrDecoderUnavailable) {
		t.Fatalf("shared registry: %v", err)
	}
	old := DefaultCompressor
	DefaultCompressor = struct{ Compressor }{writer}
	defer func() { DefaultCompressor = old }()
	if got := newTestCompressor(t, CompressorConfig{}).Compress(context.Background(), input); got.Type != TypeGzip {
		t.Fatal("default replacement affected constructor")
	}
}

func TestCompressorOwnership(t *testing.T) {
	for _, threshold := range []int{0, 2048} {
		c := newTestCompressor(t, CompressorConfig{Threshold: threshold, DisableUnframed: true})
		input := bytes.Repeat([]byte("a"), 1024)
		want := bytes.Clone(input)
		result := c.Compress(context.Background(), input)
		input[0] = 'b'
		out, _, err := c.Decompress(context.Background(), result.Data)
		if err != nil || !bytes.Equal(out, want) {
			t.Fatalf("input alias: %v", err)
		}
		if result.Type != TypeNone {
			out[0] = 'c'
			again, _, err := c.Decompress(context.Background(), result.Data)
			if err != nil || !bytes.Equal(again, want) {
				t.Fatalf("output alias: %v", err)
			}
		}
		saved := bytes.Clone(result.Data)
		for i := 0; i < 10; i++ {
			c.Compress(context.Background(), bytes.Repeat([]byte("b"), 2048))
		}
		if !bytes.Equal(saved, result.Data) {
			t.Fatal("pooled output reused")
		}
	}
}

func TestCompressorConcurrency(t *testing.T) {
	var callbacks atomic.Int64
	c := newTestCompressor(t, CompressorConfig{CompressCallback: func(context.Context, CompressionStats) { callbacks.Add(1) }})
	input := bytes.Repeat([]byte("concurrent data"), 1000)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				result := c.Compress(context.Background(), input)
				if result.Err != nil || result.Type != TypeGzip || result.Reason != "" {
					t.Errorf("unexpected compression: %s %s %v", result.Type, result.Reason, result.Err)
					return
				}
				out, typ, err := c.Decompress(context.Background(), result.Data)
				if err != nil || typ != TypeGzip || !bytes.Equal(out, input) {
					t.Errorf("round trip: %s %v", typ, err)
					return
				}
			}
		}()
	}
	wg.Wait()
	if got := callbacks.Load(); got != 320 {
		t.Fatalf("callbacks = %d", got)
	}
}

func FuzzCompressorRoundTrip(f *testing.F) {
	c := newTestCompressor(f, CompressorConfig{MaxDecodedSize: 1 << 20})
	for _, seed := range [][]byte{nil, []byte("你好"), bytes.Repeat([]byte("abc"), 1000), []byte(frameMagic)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 1<<20 {
			t.Skip()
		}
		result := c.Compress(context.Background(), input)
		if result.Err != nil {
			t.Fatal(result.Err)
		}
		out, typ, err := c.Decompress(context.Background(), result.Data)
		if err != nil || typ != result.Type || !bytes.Equal(out, input) {
			t.Fatalf("round trip: %v", err)
		}
	})
}

func FuzzCompressorDecompress(f *testing.F) {
	compatible := newTestCompressor(f, CompressorConfig{MaxDecodedSize: 64 << 10})
	strict := newTestCompressor(f, CompressorConfig{MaxDecodedSize: 64 << 10, DisableUnframed: true})
	f.Add([]byte{})
	f.Add(makeUncompressedFrame([]byte("hello")))
	f.Add(bytes.Repeat([]byte("x"), (64<<10)+1))
	f.Add(makeUncompressedFrame(bytes.Repeat([]byte("x"), (64<<10)+1)))
	f.Add(compatible.Compress(context.Background(), bytes.Repeat([]byte("a"), 1024)).Data)
	f.Add([]byte(frameMagic))
	f.Fuzz(func(t *testing.T, data []byte) {
		for _, c := range []Compressor{compatible, strict} {
			out, typ, err := c.Decompress(context.Background(), data)
			if err != nil && out != nil {
				t.Fatal("partial output on error")
			}
			if typ != TypeNone && len(out) > 64<<10 {
				t.Fatal("limit exceeded")
			}
		}
		if !bytes.HasPrefix(data, []byte(frameMagic)) {
			out, typ, err := compatible.Decompress(context.Background(), data)
			if err != nil || typ != TypeNone || !bytes.Equal(out, data) {
				t.Fatalf("unframed changed: %v", err)
			}
		}
	})
}
