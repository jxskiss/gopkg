//go:build !amd64 || go1.21

package json

import (
	"io"

	"github.com/bytedance/sonic"
)

const isSonicJIT = false

func sonicMarshalNoMapOrdering(v any) ([]byte, error) {
	return jsoniterMarshalNoMapOrdering(v)
}

func sonicMarshalNoHTMLEscape(_ sonic.API) func(v any, prefix, indent string) ([]byte, error) {
	return jsoniterMarshalNoHTMLEscape(jsoniterDefault)
}

func sonicNewEncoder(_ sonic.API) func(w io.Writer) underlyingEncoder {
	return jsoniterNewEncoder(jsoniterDefault)
}

func sonicNewDecoder(_ sonic.API) func(r io.Reader) underlyingDecoder {
	return jsoniterNewDecoder(jsoniterDefault)
}

func sonicSetEncoderDisableMapOrdering(enc *Encoder) {
}
