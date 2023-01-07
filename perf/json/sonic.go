package json

import (
	"io"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/encoder"
)

var sonicDefault = sonic.ConfigStd

func sonicMarshalNoMapOrdering(v interface{}) ([]byte, error) {
	return sonic.ConfigFastest.Marshal(v)
}

func sonicMarshalNoHTMLEscape(api sonic.API) func(v interface{}, prefix, indent string) ([]byte, error) {
	opts := api.NewEncoder(nil).(*encoder.StreamEncoder).Opts
	opts &= ^encoder.EscapeHTML
	return func(v interface{}, prefix, indent string) ([]byte, error) {
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
	if impl, ok := enc.underlyingEncoder.(*encoder.StreamEncoder); ok {
		impl.Encoder.Opts &= ^encoder.SortMapKeys
	}
}
