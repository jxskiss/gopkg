//go:build (!amd64 && !arm64) || go1.26

package bytedance_sonic

import (
	"io"

	"github.com/bytedance/sonic"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

func marshalFastest(v any) ([]byte, error) {
	return json.DefaultJSONIteratorImpl.MarshalFastest(v)
}

func marshalNoHTMLEscape(_ sonic.API) func(v any, prefix, indent string) ([]byte, error) {
	return json.StdImpl.MarshalNoHTMLEscape
}

func newEncoderFactory(_ sonic.API) func(w io.Writer) *json.Encoder {
	return json.DefaultJSONIteratorImpl.NewEncoder
}

func newDecodeFactory(_ sonic.API) func(r io.Reader) *json.Decoder {
	return json.DefaultJSONIteratorImpl.NewDecoder
}
