package json

import (
	"io"
	"runtime"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/encoder"
)

var sonicDefault = sonic.ConfigStd

func sonicMarshalNoMapOrdering(v any) ([]byte, error) {
	return sonic.ConfigFastest.Marshal(v)
}

func sonicMarshalNoHTMLEscape(api sonic.API) func(v any, prefix, indent string) ([]byte, error) {
	if isSonicFallbackImpl {
		return stdMarshalNoHTMLEscape
	}
	opts := api.NewEncoder(nil).(*encoder.StreamEncoder).Opts
	opts &= ^encoder.EscapeHTML
	return func(v any, prefix, indent string) ([]byte, error) {
		if prefix == "" && indent == "" {
			return encoder.Encode(v, opts)
		}
		return encoder.EncodeIndented(v, prefix, indent, opts)
	}
}

func sonicNewEncoder(api sonic.API) func(w io.Writer) underlyingEncoder {
	return func(w io.Writer) underlyingEncoder {
		return api.NewEncoder(w)
	}
}

func sonicNewDecoder(api sonic.API) func(r io.Reader) underlyingDecoder {
	return func(r io.Reader) underlyingDecoder {
		return api.NewDecoder(r)
	}
}

func sonicSetEncoderDisableMapOrdering(enc *Encoder) {
	if isSonicFallbackImpl {
		return
	}
	if impl, ok := enc.underlyingEncoder.(*encoder.StreamEncoder); ok {
		impl.Encoder.Opts &= ^encoder.SortMapKeys
	}
}

var isSonicFallbackImpl = testSonicImpl()

type testPanic struct{}

func (p testPanic) MarshalJSON() ([]byte, error) {
	panic("test sonic impl")
}

// Currently sonic does not provide API to tell which implementation
// is used underlying, we use this hack to check it.
// When sonic provides an API, we will switch to that API, see:
// https://github.com/bytedance/sonic/issues/367
func testSonicImpl() (isFallback bool) {
	defer func() {
		if r := recover(); r != nil {
			var pc [16]uintptr
			n := runtime.Callers(5, pc[:])
			frames := runtime.CallersFrames(pc[:n])
			for {
				f, more := frames.Next()
				funcName := f.Function
				if strings.HasPrefix(funcName, "encoding/json.(*Encoder).Encode") {
					isFallback = true
				}
				if isFallback || !more {
					break
				}
			}
		}
	}()
	sonic.Marshal(testPanic{}) //nolint:errcheck
	return
}
