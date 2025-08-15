//go:build (amd64 && go1.18 && !go1.26) || (arm64 && go1.20 && !go1.26)

package bytedance_sonic

import (
	"io"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/encoder"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

func marshalFastest(v any) ([]byte, error) {
	return sonic.ConfigFastest.Marshal(v)
}

func marshalNoHTMLEscape(api sonic.API) func(v any, prefix, indent string) ([]byte, error) {
	opts := api.NewEncoder(nil).(*encoder.StreamEncoder).Opts
	opts &= ^encoder.EscapeHTML
	return func(v any, prefix, indent string) ([]byte, error) {
		if prefix == "" && indent == "" {
			return encoder.Encode(v, opts)
		}
		return encoder.EncodeIndented(v, prefix, indent, opts)
	}
}

func newEncoderFactory(api sonic.API) func(w io.Writer) *json.Encoder {
	return func(w io.Writer) *json.Encoder {
		return &json.Encoder{UnderlyingEncoder: api.NewEncoder(w)}
	}
}

func newDecodeFactory(api sonic.API) func(r io.Reader) *json.Decoder {
	return func(r io.Reader) *json.Decoder {
		return &json.Decoder{UnderlyingDecoder: api.NewDecoder(r)}
	}
}
