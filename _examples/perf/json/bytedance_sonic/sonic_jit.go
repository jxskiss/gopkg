//go:build amd64 && !go1.22

package bytedance_sonic

import (
	"io"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/encoder"
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

func newEncoderFactory(api sonic.API) func(w io.Writer) underlyingEncoder {
	return func(w io.Writer) underlyingEncoder {
		return api.NewEncoder(w)
	}
}

func newDecodeFactory(api sonic.API) func(r io.Reader) underlyingDecoder {
	return func(r io.Reader) underlyingDecoder {
		return api.NewDecoder(r)
	}
}
