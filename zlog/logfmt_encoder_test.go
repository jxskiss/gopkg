package zlog

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func TestEncoderObjectFields(t *testing.T) {
	tests := []struct {
		desc     string
		expected string
		f        func(zapcore.Encoder)
	}{
		{"binary", `k="YWIxMg=="`, func(e zapcore.Encoder) { e.AddBinary("k", []byte("ab12")) }},
		{"bool", `k\\=true`, func(e zapcore.Encoder) { e.AddBool(`k\`, true) }}, // test key escaping once
		{"bool", `k=true`, func(e zapcore.Encoder) { e.AddBool("k", true) }},
		{"bool", `k=false`, func(e zapcore.Encoder) { e.AddBool("k", false) }},
		{"byteString", `k=v\\`, func(e zapcore.Encoder) { e.AddByteString(`k`, []byte(`v\`)) }},
		{"byteString", `k=v`, func(e zapcore.Encoder) { e.AddByteString("k", []byte("v")) }},
		{"byteString", `k=`, func(e zapcore.Encoder) { e.AddByteString("k", []byte{}) }},
		{"byteString", `k=`, func(e zapcore.Encoder) { e.AddByteString("k", nil) }},
		{"byteString", `k="a b"`, func(e zapcore.Encoder) { e.AddByteString("k", []byte("a b")) }},
		{"complex128", `k=1+2i`, func(e zapcore.Encoder) { e.AddComplex128("k", 1+2i) }},
		{"complex64", `k=1+2i`, func(e zapcore.Encoder) { e.AddComplex64("k", 1+2i) }},
		{"duration", `k=0.000000001`, func(e zapcore.Encoder) { e.AddDuration("k", 1) }},
		{"float64", `k=1`, func(e zapcore.Encoder) { e.AddFloat64("k", 1.0) }},
		{"float64", `k=10000000000`, func(e zapcore.Encoder) { e.AddFloat64("k", 1e10) }},
		{"float64", `k=NaN`, func(e zapcore.Encoder) { e.AddFloat64("k", math.NaN()) }},
		{"float64", `k=+Inf`, func(e zapcore.Encoder) { e.AddFloat64("k", math.Inf(1)) }},
		{"float64", `k=-Inf`, func(e zapcore.Encoder) { e.AddFloat64("k", math.Inf(-1)) }},
		{"float32", `k=1`, func(e zapcore.Encoder) { e.AddFloat32("k", 1.0) }},
		{"float32", `k=10000000000`, func(e zapcore.Encoder) { e.AddFloat32("k", 1e10) }},
		{"float32", `k=NaN`, func(e zapcore.Encoder) { e.AddFloat32("k", float32(math.NaN())) }},
		{"float32", `k=+Inf`, func(e zapcore.Encoder) { e.AddFloat32("k", float32(math.Inf(1))) }},
		{"float32", `k=-Inf`, func(e zapcore.Encoder) { e.AddFloat32("k", float32(math.Inf(-1))) }},
		{"int", `k=42`, func(e zapcore.Encoder) { e.AddInt("k", 42) }},
		{"int64", `k=42`, func(e zapcore.Encoder) { e.AddInt64("k", 42) }},
		{"int32", `k=42`, func(e zapcore.Encoder) { e.AddInt32("k", 42) }},
		{"int16", `k=42`, func(e zapcore.Encoder) { e.AddInt16("k", 42) }},
		{"int8", `k=42`, func(e zapcore.Encoder) { e.AddInt8("k", 42) }},
		{"string", `k=v\\`, func(e zapcore.Encoder) { e.AddString(`k`, `v\`) }},
		{"string", `k=v`, func(e zapcore.Encoder) { e.AddString("k", "v") }},
		{"string", `k=`, func(e zapcore.Encoder) { e.AddString("k", "") }},
		{"string", `k="a b"`, func(e zapcore.Encoder) { e.AddString("k", "a b") }},
		{"time", `k=1`, func(e zapcore.Encoder) { e.AddTime("k", time.Unix(1, 0)) }},
		{"uint", `k=42`, func(e zapcore.Encoder) { e.AddUint("k", 42) }},
		{"uint64", `k=42`, func(e zapcore.Encoder) { e.AddUint64("k", 42) }},
		{"uint32", `k=42`, func(e zapcore.Encoder) { e.AddUint32("k", 42) }},
		{"uint16", `k=42`, func(e zapcore.Encoder) { e.AddUint16("k", 42) }},
		{"uint8", `k=42`, func(e zapcore.Encoder) { e.AddUint8("k", 42) }},
		{"uintptr", `k=42`, func(e zapcore.Encoder) { e.AddUintptr("k", 42) }},
		{
			desc:     "array (success)",
			expected: `k=a,b`,
			f: func(e zapcore.Encoder) {
				assert.NoError(
					t,
					e.AddArray(`k`, zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
						for _, s := range []string{"a", "b"} {
							enc.AppendString(s)
						}
						return nil
					}),
					),
					"Unexpected error calling MarshalLogArray.",
				)
			},
		},
		{
			desc:     "namespace",
			expected: `outermost.outer.foo=1 outermost.outer.inner.foo=2`,
			f: func(e zapcore.Encoder) {
				e.OpenNamespace("outermost")
				e.OpenNamespace("outer")
				e.AddInt("foo", 1)
				e.OpenNamespace("inner")
				e.AddInt("foo", 2)
				e.OpenNamespace("innermost")
			},
		},
	}

	for _, tt := range tests {
		assertOutput(t, tt.desc, tt.expected, tt.f)
	}
}

func assertOutput(t testing.TB, desc string, expected string, f func(zapcore.Encoder)) {
	enc := &logfmtEncoder{buf: bufferpool.Get(), EncoderConfig: &zapcore.EncoderConfig{
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}}
	f(enc)
	assert.Equal(t, expected, enc.buf.String(), "Unexpected encoder output after adding a %s.", desc)

	enc.truncate()
	enc.AddString("foo", "bar")
	f(enc)
	expectedPrefix := `foo=bar`
	if expected != "" {
		// If we expect output, it should be comma-separated from the previous
		// field.
		expectedPrefix += " "
	}
	assert.Equal(t, expectedPrefix+expected, enc.buf.String(), "Unexpected encoder output after adding a %s as a second field.", desc)
}

func TestEncodeCaller(t *testing.T) {
	enc := &logfmtEncoder{buf: bufferpool.Get(), EncoderConfig: &zapcore.EncoderConfig{
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}}

	var buf *buffer.Buffer
	var err error
	encodeEntry := func() {
		buf, err = enc.EncodeEntry(
			zapcore.Entry{
				Level:      zapcore.DebugLevel,
				Time:       time.Time{},
				LoggerName: "test",
				Message:    "caller test",
				Caller: zapcore.EntryCaller{
					Defined: true,
					File:    "h2g2.go",
					Line:    42,
				},
			},
			[]zapcore.Field{
				zap.String("k", "v"),
			},
		)
	}

	encodeEntry()
	assert.Nil(t, err)
	assert.Equal(t, "k=v\n", buf.String())

	enc.truncate()
	enc.EncoderConfig.CallerKey = "caller"
	encodeEntry()
	assert.Nil(t, err)
	assert.Equal(t, "caller=h2g2.go:42 k=v\n", buf.String())
}

func TestEncodeStacktrace(t *testing.T) {
	enc := &logfmtEncoder{buf: bufferpool.Get(), EncoderConfig: &zapcore.EncoderConfig{
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}}

	var buf *buffer.Buffer
	var err error
	encodeEntry := func() {
		buf, err = enc.EncodeEntry(
			zapcore.Entry{
				Level:      zapcore.DebugLevel,
				Time:       time.Time{},
				LoggerName: "test",
				Message:    "stacktrace test",
				Stack: `panic: an unexpected error occurred

goroutine 1 [running]:
main.main()
		/go/src/github.com/jsternberg/myawesomeproject/h2g2.go:4 +0x39
`,
			},
			[]zapcore.Field{
				zap.String("k", "v"),
			},
		)
	}

	encodeEntry()
	assert.Nil(t, err)
	assert.Equal(t, "k=v\n", buf.String())

	enc.truncate()
	enc.EncoderConfig.StacktraceKey = "stacktrace"
	encodeEntry()
	assert.Nil(t, err)
	assert.Equal(t, `k=v stacktrace="panic: an unexpected error occurred\n\ngoroutine 1 [running]:\nmain.main()\n\t\t/go/src/github.com/jsternberg/myawesomeproject/h2g2.go:4 +0x39\n"
`, buf.String())
}
