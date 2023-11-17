//go:build !amd64 || go1.22

package json

import (
	"io"

	"github.com/bytedance/sonic"
)

const supportSonicJIT = false

func sonicMarshalFastest(v any) ([]byte, error) {
	return jsoniterMarshalFastest(v)
}

func sonicMarshalNoHTMLEscape(_ sonic.API) func(v any, prefix, indent string) ([]byte, error) {
	return stdMarshalNoHTMLEscape
}

func sonicNewEncoder(_ sonic.API) func(w io.Writer) underlyingEncoder {
	return jsoniterNewEncoder(jsoniterDefault)
}

func sonicNewDecoder(_ sonic.API) func(r io.Reader) underlyingDecoder {
	return jsoniterNewDecoder(jsoniterDefault)
}

func sonicSetEncoderDisableMapOrdering(enc *Encoder) {
}
